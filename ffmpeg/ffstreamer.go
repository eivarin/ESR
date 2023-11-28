package ffmpeg

import (
	"errors"
	"fmt"
	"log"
	"main/ffmpeg/OS"
	"main/ffmpeg/Shared"
	"net"
	"sync"
)

type streamInstance struct {
	name string
	sdp_string string
	ffmpeg *OS.Streamer
	conns []*net.UDPConn
}

type FFStreamer struct {
	streamer_name string
	instances map[string]*streamInstance
	lock sync.RWMutex
	remoteCon *net.UDPConn
	logger *log.Logger
	basePort int
	debug bool
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


func (ffs *FFStreamer) AddStream(path string) {
	ports := shared.GetPorts(ffs.basePort)
	sdp_string := OS.Generate_SDP_From_FFMPEG(path)
	ffs.lock.Lock()
	defer ffs.lock.Unlock()
	new_file_name := ffs.streamer_name + "." + path
	ffs.instances[new_file_name] = new_streamInstance(new_file_name, path, sdp_string, ports, ffs.logger)
	for i := 0; i < 4; i++ {
		go func (name string, i int)  {
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
				encoded_packet := encodeMediaPacket(MediaPacket{path, byte(i), n, buff[:n]})
				ffs.remoteCon.Write(encoded_packet)
			}
		}(new_file_name, i)
	}
	ffs.logger.Printf("Added stream %s\n", new_file_name)
}

func (ffs *FFStreamer) RemoveStream(file_name string) {
	ffs.lock.Lock()
	defer ffs.lock.Unlock()
	if _, ok := ffs.instances[file_name]; ok {
		ffs.instances[file_name].stop(ffs.logger)
		delete(ffs.instances, file_name)
	}
}

func new_streamInstance(name string, media_path string, SDP_string string, ports []int, logger *log.Logger) *streamInstance {
	si := new(streamInstance)
	si.sdp_string = SDP_string
	si.name = name
	si.conns = make([]*net.UDPConn, 4)
	si.ffmpeg = OS.NewStreamer(media_path, logger, false, ports)
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
	if _, ok := ffs.instances[file_name]; ok {
		return ffs.instances[file_name].sdp_string, nil
	} else {
		return "", errors.New("No stream with name " + file_name)
	}
}