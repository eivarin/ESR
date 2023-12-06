package main

import (
	"io"
	"log"
	"main/ffmpeg"
	"main/overlay"
	"main/packets"
	"main/packets/join"
	"main/packets/ready"
	"main/packets/requeststream"
	"main/safesockets"
	"main/shell"
	"net"
	"os"
	"time"
)

type OverlayServer struct {
	logger *log.Logger
	overlay *overlay.Overlay
	rpConn *safesockets.SafeTCPConn
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
	rpConn, _ := net.DialTCP("tcp", nil, rpAddr)
	server.rpConn = safesockets.NewSafeTCPConn(rpConn)
	server.rpConn.SafeWrite(jp)
	p, _, _, err := server.rpConn.SafeRead()
	if err != nil {
		log.Panic("JoinOverlayRPResponsePacket failed", err)
	}
	
	rjoin := p.(*join.JoinOverlayRPResponsePacket)
	if !rjoin.Success {
		log.Panic("JoinOverlayRPResponsePacket failed")
	}
	server.publishStreams()
	go func() {
		for {
			p, t, _, err := server.rpConn.SafeRead()
			if err != nil {
				if err == io.EOF {
					log.Fatal("RP connection closed")
				} else {
					server.logger.Println("RP connection error", err)
				}
			}
			server.HandlePacketFromRP(p, t, server.rpConn)
		}
	}()
}

func (s *OverlayServer) LeaveOverlay() {
	
}

func (s *OverlayServer) HandlePacketFromNeighbor(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn) {
	
}

func (s *OverlayServer) HandlePacketFromRP(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn) {
	switch t {
	case packets.OverlayPacketTypeRequestStream:
		rsp := p.(*requeststream.RequestStreamPacket)
		s.logger.Println("Requesting stream", s.availableStreams[rsp.StreamName].Path)
		sdp, _ := s.ffstreamer.GetSDP(s.availableStreams[rsp.StreamName].Path)
		resp := requeststream.NewRequestStreamResponsePacket(true, rsp.StreamName, sdp)
		s.logger.Println(resp)
		conn.SafeWrite(resp)
	case packets.OverlayPacketTypeReady:
		ready := p.(*ready.ReadyPacket)
		s.logger.Println("Starting stream", ready.StreamName)
		s.availableStreams[ready.StreamName].Running = true
		s.ffstreamer.AddStream(s.availableStreams[ready.StreamName].Path, ready.StreamName)
	}
	
}


func main() {
	Neighbors, name := shell.GetRealArgs("File path with Server Configs", "Name of the config to read from static/configs/S${NAME}.conf")
	logger := log.New(os.Stdout, "", log.LstdFlags)
	s := shell.NewShell(logger, name)

	overlayNode := NewOverlayServer(logger, Neighbors[0], Neighbors[1:2], Neighbors[2], s)
	overlayNode.registerCommands(s)
	s.Run()
}