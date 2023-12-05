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

type OverlayNode struct {
	logger *log.Logger
	overlay *overlay.Overlay
	rpConn *net.TCPConn
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
	packet := packets.NewOverlayPacket(jp)
	n.rpConn, _ = net.DialTCP("tcp", nil, rpAddr)
	n.rpConn.Write(packet.Encode())
	buf := make([]byte, 1500)
	_, err := n.rpConn.Read(buf)
	if err != nil {
		log.Panic("JoinOverlayRPResponsePacket failed", err)
	}
	
	rPacket := packets.OverlayPacket{}
	rPacket.Decode(buf)
	rjoin := rPacket.InnerPacket().(*join.JoinOverlayRPResponsePacket)
	if !rjoin.Success {
		log.Panic("JoinOverlayRPResponsePacket failed")
	}
	go func() {
		for {
			buf := make([]byte, 1500)
			_, err := n.rpConn.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Fatal("RP connection closed")
				} else {
					n.logger.Println("RP connection error", err)
				}
			}
			n.HandlePacketFromRP(buf, n.rpConn)
		}
	}()
}

func (n *OverlayNode) LeaveOverlay() {
	
}

func (n *OverlayNode) HandlePacketFromNeighbor(p *packets.OverlayPacket) {
	
}

func (n *OverlayNode) HandlePacketFromRP(buf []byte, conn *net.TCPConn) {
	p := packets.OverlayPacket{}
	p.Decode(buf)
}

func main() {
	Neighbors, name := shell.GetRealArgs("File path with node configs", "Name of the config to read from static/configs/O${NAME}.conf")
	logger := log.New(os.Stdout, "", log.LstdFlags)
	s := shell.NewShell(logger, name)

	overlayNode := NewOverlayNode(logger, Neighbors[0], Neighbors[1:], s)
	overlayNode.registerCommands(s)
	s.Run()
}