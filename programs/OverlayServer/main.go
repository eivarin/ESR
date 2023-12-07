package main

import (
	"io"
	"log"
	"main/ffmpeg"
	"main/overlay"
	"main/packets"
	"main/packets/join"
	"main/packets/nolonguer"
	"main/packets/requeststream"
	"main/packets/rtt"
	"main/packets/startstream"
	"main/safesockets"
	"main/shell"
	"net"
	"os"
	"time"
)

type OverlayServer struct {
	logger           *log.Logger
	overlay          *overlay.Overlay
	rpConn           *safesockets.SafeTCPConn
	ffstreamer       *ffmpeg.FFStreamer
	availableStreams map[string]*availableStream
	localAddr        string
	sh               *shell.Shell
}

func NewOverlayServer(logger *log.Logger, localAddr string, neighborsAddrs []string, manifestPath string, sh *shell.Shell) *OverlayServer {
	oServer := new(OverlayServer)
	oServer.sh = sh
	oServer.logger = logger
	oServer.localAddr = localAddr
	oServer.availableStreams = make(map[string]*availableStream)
	oServer.populateFromManifest(manifestPath)
	oServer.overlay = overlay.NewOverlay(logger, 5000, overlay.NodeServer, localAddr, neighborsAddrs, oServer)
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
	s.logger.Println("Received packet from RP with type", t)
	switch t {
	case packets.OverlayPacketTypeRequestStream:
		rsp := p.(*requeststream.RequestStreamPacket)
		s.logger.Println("Requesting stream", s.availableStreams[rsp.StreamName].Path)
		sdp, _ := s.ffstreamer.GetSDP(s.availableStreams[rsp.StreamName].Path)
		resp := requeststream.NewRequestStreamResponsePacket(true, rsp.StreamName, sdp)
		conn.SafeWrite(resp)
	case packets.OverlayPacketTypeStartStream:
		ssp := p.(*startstream.StartStreamPacket)
		s.logger.Println("Starting stream", ssp.StreamName)
		s.availableStreams[ssp.StreamName].Running = true
		s.ffstreamer.AddStream(s.availableStreams[ssp.StreamName].Path, ssp.StreamName, ssp.StartTimeSeconds)
	case packets.OverlayPacketTypeRtt:
		rttP := p.(*rtt.RttPacket)
		s.logger.Println("RTT request from RF: ", rttP.SenderAddr)
		s.overlay.RttLock.Lock()
		lad, _ := net.ResolveUDPAddr("udp", s.localAddr+":4994")
		lConn, _ := net.ListenUDP("udp", lad)
		rad, _ := net.ResolveUDPAddr("udp", rttP.SenderAddr+":4994")
		rConn, _ := net.DialUDP("udp", nil, rad)

		go func() {
			buf := make([]byte, 10)
			lConn.Read(buf)
			rConn.Write([]byte("rtt"))
			s.overlay.RttLock.Unlock()
			rConn.Close()
			lConn.Close()
		}()
		respRtt := rtt.NewRttResponsePacket(s.localAddr)
		conn.SafeWrite(respRtt)
	case packets.OverlayPacketTypeNoLonguerInterested:
		nli := p.(*nolonguer.NoLonguerInterestedPacket)
		s.logger.Println("No longuer interested in stream", nli.StreamName)
		s.availableStreams[nli.StreamName].Running = false
		s.logger.Println("Stopping stream", nli.StreamName)
		s.ffstreamer.RemoveStream(nli.StreamName)
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
