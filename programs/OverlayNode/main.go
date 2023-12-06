package main

import (
	"io"
	"log"
	"main/ffmpeg"
	"main/overlay"
	"main/packets"
	"main/packets/join"
	"main/packets/redirects"
	"main/safesockets"
	"main/shell"
	"net"
	"os"
	"time"
)

type OverlayNode struct {
	logger *log.Logger
	overlay *overlay.Overlay
	rpConn *safesockets.SafeTCPConn
	sh *shell.Shell
	redirecter *ffmpeg.FFRedirecter
}

func NewOverlayNode(logger *log.Logger, localAddr string, neighborsAddrs []string, sh *shell.Shell) *OverlayNode {
	node := new(OverlayNode)
	node.logger = logger
	node.overlay = overlay.NewOverlay(logger, 5000, overlay.NodeOverlay, localAddr, neighborsAddrs, node)
	udpAddr, _:= net.ResolveUDPAddr("udp", ":4999")
	node.redirecter = ffmpeg.NewFFRedirecter(logger, udpAddr, false)
	node.sh = sh
	return node
}

func (n *OverlayNode) JoinOverlay(rpAddr *net.TCPAddr, localAddr string, neighborsAddrs []string, neighborsRTTs map[string]time.Duration) {

	jp := join.NewJoinOverlayRPPacket(overlay.NodeOverlay, localAddr, neighborsAddrs, neighborsRTTs)
	rpConn, _ := net.DialTCP("tcp", nil, rpAddr)
	n.rpConn = safesockets.NewSafeTCPConn(rpConn)
	n.rpConn.SafeWrite(jp)
	p, _, _, err := n.rpConn.SafeRead()
	if err != nil {
		log.Panic("JoinOverlayRPResponsePacket failed", err)
	}
	rjoin := p.(*join.JoinOverlayRPResponsePacket)
	if !rjoin.Success {
		log.Panic("JoinOverlayRPResponsePacket failed")
	}
	go func() {
		for {
			p, t, _, err := n.rpConn.SafeRead()
			if err != nil {
				if err == io.EOF {
					log.Fatal("RP connection closed")
				} else {
					n.logger.Println("RP connection error", err)
				}
			}
			n.HandlePacketFromRP(p, t, n.rpConn)
		}
	}()
}

func (n *OverlayNode) LeaveOverlay() {
	
}

func (n *OverlayNode) HandlePacketFromNeighbor(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn) {
	
}

func (n *OverlayNode) HandlePacketFromRP(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn) {
	switch t {
	case packets.OverlayPacketTypeRedirects:
		rp := p.(*redirects.RedirectsPacket)
		n.logger.Println("Received RedirectsPacket", rp)
		mappings := make(map[string]*safesockets.SafeUDPWritter)
		for _, r := range rp.Redirects {
			addr, _ := net.ResolveUDPAddr("udp", r+":4999")
			conn, _ := net.DialUDP("udp", nil, addr)
			mappings[r] = safesockets.NewSafeUDPWritter(conn)
		}
		n.redirecter.AddStream(rp.StreamName, mappings)
	}
}

func main() {
	Neighbors, name := shell.GetRealArgs("File path with node configs", "Name of the config to read from static/configs/O${NAME}.conf")
	logger := log.New(os.Stdout, "", log.LstdFlags)
	s := shell.NewShell(logger, name)

	overlayNode := NewOverlayNode(logger, Neighbors[0], Neighbors[1:], s)
	overlayNode.registerCommands(s)
	s.Run()
}