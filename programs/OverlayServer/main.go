package main

import (
	"io"
	"log"
	"main/ffmpeg"
	"main/overlay"
	"main/packets"
	"main/packets/join"
	"main/shell"
	"net"
	"os"
	"time"
)

type OverlayServer struct {
	logger *log.Logger
	overlay *overlay.Overlay
	rpConn *net.TCPConn
	ffstreamer *ffmpeg.FFStreamer
	availableStreams map[string]*availableStream
	localAddr string
	sh *shell.Shell
}

func NewOverlayServer(logger *log.Logger, localAddr string, neighborsAddrs []string, manifestPath string, sh *shell.Shell) *OverlayServer {
	oServer := new(OverlayServer)
	oServer.sh = sh
	oServer.logger = logger
	oServer.localAddr = localAddr
	oServer.availableStreams = make(map[string]*availableStream)
	oServer.populateFromManifest(manifestPath)
	oServer.overlay = overlay.NewOverlay(logger,  5000, overlay.NodeServer, localAddr, neighborsAddrs, oServer)
	oServer.ffstreamer = ffmpeg.NewFFStreamer(5000, logger, localAddr, oServer.overlay.RPAddr, false)
	return oServer
}

func (server *OverlayServer) JoinOverlay(rpAddr *net.TCPAddr, localAddr string, neighborsAddrs []string, neighborsRTTs map[string]time.Duration) {

	jp := join.NewJoinOverlayRPPacket(overlay.NodeServer, localAddr, neighborsAddrs, neighborsRTTs)
	packet := packets.NewOverlayPacket(jp)
	server.rpConn, _ = net.DialTCP("tcp", nil, rpAddr)
	server.rpConn.Write(packet.Encode())
	buf := make([]byte, 1500)
	_, err := server.rpConn.Read(buf)
	if err != nil {
		log.Panic("JoinOverlayRPResponsePacket failed", err)
	}
	
	rPacket := packets.OverlayPacket{}
	rPacket.Decode(buf)
	rjoin := rPacket.InnerPacket().(*join.JoinOverlayRPResponsePacket)
	if !rjoin.Success {
		log.Panic("JoinOverlayRPResponsePacket failed")
	}
	server.publishStreams()
	go func() {
		for {
			buf := make([]byte, 1500)
			_, err := server.rpConn.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Fatal("RP connection closed")
				} else {
					server.logger.Println("RP connection error", err)
				}
			}
			server.HandlePacketFromRP(buf, server.rpConn)
		}
	}()
}

func (s *OverlayServer) LeaveOverlay() {
	
}

func (s *OverlayServer) HandlePacketFromNeighbor(p *packets.OverlayPacket) {
	
}

func (s *OverlayServer) HandlePacketFromRP(buf []byte, conn *net.TCPConn) {
	p := packets.OverlayPacket{}
	p.Decode(buf)
}


func main() {
	Neighbors, name := shell.GetRealArgs("File path with Server Configs", "Name of the config to read from static/configs/S${NAME}.conf")
	logger := log.New(os.Stdout, "", log.LstdFlags)
	s := shell.NewShell(logger, name)

	overlayNode := NewOverlayServer(logger, Neighbors[0], Neighbors[1:2], Neighbors[2], s)
	overlayNode.registerCommands(s)
	s.Run()
}