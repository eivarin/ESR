package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"main/ffmpeg/OS"
	"main/ffmpeg/Shared"
	"main/shell"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pion/sdp"
)

type clientInstance struct {
	name string
	ffplay *OS.Player
	ports []int
	localCons []*net.UDPConn
}

type FFClient struct {
	instances map[string]*clientInstance
	lock sync.RWMutex
	remoteCon *net.UDPConn
	logger *log.Logger
	basePort int
	debug bool
}

func new_clientInstance(name string, SDP_string string, ports []int, logger *log.Logger, debug bool) *clientInstance {
	ci := new(clientInstance)
	ci.name = name
	ci.ports = ports
	write_SDP_to_file(SDP_string, "./"+name+".sdp", ports[0], ports[2])
	ci.ffplay = OS.NewPlayer(name+".sdp", logger, debug)
	ci.localCons = make([]*net.UDPConn, 4)
	for i := 0; i < 4; i++ {
		localAddr, _ := net.ResolveUDPAddr("udp", ":"+fmt.Sprint(ports[i]))
		conn, err := net.DialUDP("udp", nil, localAddr)
		if err != nil {
			fmt.Println(err)
		}
		ci.localCons[i] = conn
	}
	return ci
}

func NewFFClient(basePort int, logger *log.Logger, debug bool) *FFClient {
	ffc := new(FFClient)
	ffc.instances = make(map[string]*clientInstance)
	ffc.logger = logger
	ffc.basePort = basePort
	ffc.debug = debug
	remoteAddr, _ := net.ResolveUDPAddr("udp", ":"+fmt.Sprint(basePort-1))
	conn, err := net.ListenUDP("udp", remoteAddr)
	if err != nil {
		fmt.Println(err)
	}
	ffc.remoteCon = conn
	go ffc.recieveUDPPackets()
	return ffc
}

func (ffc *FFClient) Stop() {
	ffc.remoteCon.Close()

	for _, instance := range ffc.instances {
		instance.ffplay.Stop()
		for i := 0; i < 4; i++ {
			instance.localCons[i].Close()
		}
		os.Remove("./"+instance.name+".sdp")
	}
}

func (ffc *FFClient) recieveUDPPackets() {
	for {
		buf := make([]byte, 2000)
		n, err := ffc.remoteCon.Read(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			} else {
				ffc.logger.Print(err)
			}
		}
		go func(n int) {
			parsed_packet := decodeMediaPacket(buf[:n])
			ffc.lock.RLock()
			defer ffc.lock.RUnlock()
			if _, ok := ffc.instances[parsed_packet.StreamName]; ok {
				ffc.instances[parsed_packet.StreamName].localCons[parsed_packet.Type].Write(parsed_packet.Payload[:parsed_packet.PayloadSize])
			} else if ffc.debug {
				bs, _ := json.Marshal(ffc.instances)
				fmt.Println(string(bs))
				ffc.logger.Printf("Recieved unrequested stream: %s\n", parsed_packet.StreamName)
			}
		}(n)
	}
}


func (ffc *FFClient) AddInstance(name string, SDP_string string, debug bool) {
	ffc.lock.Lock()
	defer ffc.lock.Unlock()
	if _, ok := ffc.instances[name]; !ok {
		ports := shared.GetPorts(ffc.basePort)
		ci := new_clientInstance(name, SDP_string, ports, ffc.logger, debug)
		ffc.instances[name] = ci
	}
	time.Sleep(1 * time.Second)
}

func (ffc *FFClient) RemoveInstance(name string) {
	ffc.lock.Lock()
	defer ffc.lock.Unlock()
	if _, ok := ffc.instances[name]; ok {
		ffc.instances[name].ffplay.Stop()
		for i := 0; i < 4; i++ {
			ffc.instances[name].localCons[i].Close()
		}
		os.Remove("./"+name+".sdp")
		delete(ffc.instances, name)
	}
}


func write_SDP_to_file(SDP_string string, output_file string, videoPort int, audioPort int) {
	file, _ := os.Create(output_file)
	Parsed_SDP := sdp.SessionDescription{}
	Parsed_SDP.Unmarshal(SDP_string)
	
	if Parsed_SDP.MediaDescriptions[0].MediaName.Media == "video"{
		Parsed_SDP.MediaDescriptions[0].MediaName.Port.Value = videoPort
		Parsed_SDP.MediaDescriptions[1].MediaName.Port.Value = audioPort
	} else {
		Parsed_SDP.MediaDescriptions[0].MediaName.Port.Value = audioPort
		Parsed_SDP.MediaDescriptions[1].MediaName.Port.Value = videoPort
	}
	corrected_SDP := Parsed_SDP.Marshal()
	file.WriteString(corrected_SDP)
	file.Close()
}

//Possible commands:
//	new <instance_name> <sdp_path> <debug:optional>
//	del <instance_name>
//	gen <path> <stream_name>
// 	list
//  help
func (ffc *FFClient) RegisterClientCommands(shell *shell.Shell) {
	shell.RegisterCommand("new", ffc.ShellAddInstance)
	shell.RegisterCommand("del", ffc.ShellRemoveInstance)
	shell.RegisterCommand("gen", ffc.ShellGenerateSDPFile)
	shell.RegisterCommand("list", ffc.ShellListInstances)
	shell.RegisterCommand("help", shellHelp)
}

//Possible args:
//	<instance_name> Mandatory
//	<sdp_path> Mandatory
//	<debug> Optional default false

func (ffc *FFClient) ShellAddInstance(args []string) {
	debug := false
	if len(args) < 2 {
		ffc.logger.Println("Usage: new <instance_name> <sdp_path> <debug:optional>")
		return
	} else if len(args) == 3 {
		debug, _ = strconv.ParseBool(args[2])
	}

	// Access the parsed values
	name := args[0]
	sdpPath := args[1]

	// Read the SDP file
	sdpBytes, err := os.ReadFile(sdpPath)
	if err != nil {
		ffc.logger.Println("Error reading SDP file:", err)
		return
	}
	sdp := string(sdpBytes)

	// Use the parsed values
	ffc.AddInstance(name, sdp, debug)
}

//Possible args:
//	<instance_name> Mandatory
func (ffc *FFClient) ShellRemoveInstance(args []string) {
	if len(args)  != 1{
		ffc.logger.Println("Usage: del <instance_name>")
		return
	}

	// Access the parsed values
	name := args[0]

	ffc.RemoveInstance(name)
}

//Possible args:
//	<path> Mandatory
// <stream_name> Mandatory
func (ffc *FFClient) ShellGenerateSDPFile(args []string) {
	if len(args)  != 2{
		ffc.logger.Println("Usage: gen <path>")
		return
	}

	// Access the parsed values
	path := args[0]
	sdpPath := args[1] + ".sdp"

	// Generate the SDP file
	sdp := OS.Generate_SDP_From_FFMPEG(path)
	write_SDP_to_file(sdp, sdpPath, 5000, 5002)
	ffc.logger.Println("SDP file written to", sdpPath)
}

func (ffc *FFClient) ShellListInstances(args []string) {
	ffc.lock.RLock()
	defer ffc.lock.RUnlock()
	for name := range ffc.instances {
		fmt.Println(name)
	}
}

func shellHelp(args []string) {
	fmt.Println("Possible commands:")
	fmt.Println("	new <instance_name> <sdp_path> <debug:optional>")
	fmt.Println("	del <instance_name>")
	fmt.Println("	gen <path> <stream_name>")
	fmt.Println("	help")
}