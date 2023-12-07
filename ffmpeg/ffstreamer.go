package ffmpeg

import (
	"errors"
	"fmt"
	"log"
	"main/ffmpeg/OS"
	"main/ffmpeg/Shared"
	"main/shell"
	"net"
	"sync"
)

type streamInstance struct {
	name       string
	sdp_string string
	ffmpeg     *OS.Streamer
	conns      []*net.UDPConn
}

type FFStreamer struct {
	streamer_name string
	instances     map[string]*streamInstance
	lock          sync.RWMutex
	remoteCon     *net.UDPConn
	logger        *log.Logger
	basePort      int
	debug         bool
}

func NewFFStreamer(basePort int, logger *log.Logger, streamer_name string, rp string, debug bool) *FFStreamer {
	ffs := new(FFStreamer)
	ffs.debug = debug
	ffs.basePort = basePort
	ffs.streamer_name = streamer_name
	ffs.instances = make(map[string]*streamInstance)
	ffs.logger = logger
	logger.Println(rp)
	remoteAddr, _ := net.ResolveUDPAddr("udp", rp+":"+fmt.Sprint(basePort-1))
	conn, _ := net.DialUDP("udp", nil, remoteAddr)
	ffs.remoteCon = conn
	return ffs
}

func (ffs *FFStreamer) AddStream(path string, name string, startAtSeconds int64) {
	ports := shared.GetPorts(ffs.basePort)
	sdp_string := OS.Generate_SDP_From_FFMPEG(path)
	ffs.lock.Lock()
	defer ffs.lock.Unlock()
	ffs.instances[name] = new_streamInstance(name, path, sdp_string, ports, startAtSeconds, ffs.logger)
	for i := 0; i < 4; i++ {
		go func(name string, i int) {
			for {
				buff := make([]byte, 2000)
				n, _, err := ffs.instances[name].conns[i].ReadFromUDP(buff)
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						break
					} else {
						ffs.logger.Fatal(err)
					}
				}
				encoded_packet := encodeMediaPacket(MediaPacket{name, byte(i), n, buff[:n]})
				ffs.remoteCon.Write(encoded_packet)
			}
		}(name, i)
	}
	ffs.logger.Printf("Added stream [%s] from path: %s\n", name, path)
}

func (ffs *FFStreamer) RemoveStream(file_name string) {
	ffs.lock.Lock()
	defer ffs.lock.Unlock()
	if _, ok := ffs.instances[file_name]; ok {
		ffs.instances[file_name].stop(ffs.logger)
		delete(ffs.instances, file_name)
	}
}

func new_streamInstance(name string, media_path string, SDP_string string, ports []int, startAtSeconds int64, logger *log.Logger) *streamInstance {
	si := new(streamInstance)
	si.sdp_string = SDP_string
	si.name = name
	si.conns = make([]*net.UDPConn, 4)
	si.ffmpeg = OS.NewStreamer(media_path, logger, false, ports, startAtSeconds)
	for i := 0; i < 4; i++ {
		remoteAddr, _ := net.ResolveUDPAddr("udp", ":"+fmt.Sprint(ports[i]))
		conn, err := net.ListenUDP("udp", remoteAddr)
		if err != nil {
			fmt.Println(err)
		}
		si.conns[i] = conn
	}
	return si
}

func (ffs *FFStreamer) Stop() {
	ffs.logger.Print("Stopping FFStreamer")
	ffs.remoteCon.Close()
	for _, instance := range ffs.instances {
		instance.stop(ffs.logger)
	}
}

func (streamInstance *streamInstance) stop(logger *log.Logger) {
	logger.Printf("Stopping stream %s\n", streamInstance.name)
	streamInstance.ffmpeg.Stop()
	for i := 0; i < 4; i++ {
		streamInstance.conns[i].Close()
	}
}

func (ffs *FFStreamer) GetSDP(file_name string) (string, error) {
	ffs.lock.RLock()
	defer ffs.lock.RUnlock()
	return OS.Generate_SDP_From_FFMPEG(file_name), nil
}

// Possible commands:
// new <file_name> <stream_name>
// del <stream_name>
// get <stream_name>
// help
func (ffs *FFStreamer) RegisterCommands(shell *shell.Shell) {
	shell.RegisterCommand("new", ffs.ShellAddStream)
	shell.RegisterCommand("del", ffs.ShellRemoveStream)
	shell.RegisterCommand("get", ffs.ShellGetSDP)
	shell.RegisterCommand("help", ffs.ShellHelp)
}

func (ffs *FFStreamer) ShellAddStream(args []string) {
	if len(args) != 2 {
		ffs.logger.Println("Usage: new <file_name> <stream_name>")
		return
	}
	file_name := args[0]
	stream_name := args[1]
	ffs.AddStream(file_name, stream_name, 0)
}

func (ffs *FFStreamer) ShellRemoveStream(args []string) {
	if len(args) != 1 {
		ffs.logger.Println("Usage: del <stream_name>")
		return
	}
	stream_name := args[0]
	ffs.RemoveStream(stream_name)
}

func (ffs *FFStreamer) ShellGetSDP(args []string) {
	if len(args) != 1 {
		ffs.logger.Println("Usage: get <stream_name>")
		return
	}
	stream_name := args[0]
	sdp, err := ffs.GetSDP(stream_name)
	if err != nil {
		ffs.logger.Println(err)
		return
	}
	ffs.logger.Printf("SDP for stream %s:\n%s\n", stream_name, sdp)
}

func (ffs *FFStreamer) ShellHelp(args []string) {
	ffs.logger.Println("Possible commands:")
	ffs.logger.Println("	new <file_name> <stream_name>")
	ffs.logger.Println("	del <stream_name>")
	ffs.logger.Println("	get <stream_name>")
	ffs.logger.Println("	help")
}
