package main

import (
	"io"
	"log"
	"main/ffmpeg"
	"main/overlay"
	"main/packets"
	"main/packets/getstreams"
	"main/packets/join"
	"main/packets/ready"
	"main/packets/requeststream"
	"main/safesockets"
	"main/shell"
	"net"
	"os"
	"time"
)

type OverlayClient struct {
	logger   *log.Logger
	overlay  *overlay.Overlay
	rpConn   *safesockets.SafeTCPConn
	ffplayer *ffmpeg.FFClient
	streams  map[string]bool
	sh       *shell.Shell
}

func NewOverlayClient(logger *log.Logger, localAddr string, neighborsAddrs []string, sh *shell.Shell) *OverlayClient {
	oClient := new(OverlayClient)
	oClient.logger = logger
	oClient.overlay = overlay.NewOverlay(logger, 5000, overlay.NodeClient, localAddr, neighborsAddrs, oClient)
	oClient.ffplayer = ffmpeg.NewFFClient(5000, logger, false)
	oClient.streams = make(map[string]bool)
	oClient.sh = sh
	return oClient
}

func (c *OverlayClient) JoinOverlay(rpAddr *net.TCPAddr, localAddr string, neighborsAddrs []string, neighborsRTTs map[string]time.Duration) {
	jp := join.NewJoinOverlayRPPacket(overlay.NodeClient, localAddr, neighborsAddrs, neighborsRTTs)
	rpConn, _ := net.DialTCP("tcp", nil, rpAddr)
	c.rpConn = safesockets.NewSafeTCPConn(rpConn)
	c.rpConn.SafeWrite(jp)
	p, _, _, err := c.rpConn.SafeRead()
	if err != nil {
		log.Panic("JoinOverlayRPResponsePacket failed", err)
	}
	rjoin := p.(*join.JoinOverlayRPResponsePacket)
	if !rjoin.Success {
		log.Panic("JoinOverlayRPResponsePacket failed")
	}
	go func() {
		for {
			p, t, _, err := c.rpConn.SafeRead()
			if err != nil {
				if err == io.EOF {
					log.Fatal("RP connection closed")
				} else {
					c.logger.Println("RP connection error", err)
				}
			}
			c.HandlePacketFromRP(p, t, c.rpConn)
		}
	}()
}

func (c *OverlayClient) LeaveOverlay() {

}

func (c *OverlayClient) HandlePacketFromNeighbor(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn) {

}

func (c *OverlayClient) HandlePacketFromRP(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn) {
	switch t {
	case packets.OverlayPacketTypeGetStreamsResponse:
		gsp, _ := p.(*getstreams.GetStreamsResponsePacket)
		c.streams = make(map[string]bool)
		for _, stream := range gsp.StreamNames {
			c.streams[stream] = true
			c.logger.Printf("%s \n", stream)
		}
		c.sh.Lock.Unlock()
	case packets.OverlayPacketTypeRequestStreamResponse:
		rsp, _ := p.(*requeststream.RequestStreamResponsePacket)
		c.logger.Println(rsp)
		if !rsp.Status {
			c.logger.Printf("Stream %s not found \n", rsp.StreamName)
		} else {
			c.ffplayer.AddInstance(rsp.StreamName, rsp.Sdp, true)
			c.logger.Printf("Playing %s \n", rsp.StreamName)
			readyP := ready.NewReadyPacket(rsp.StreamName)
			c.rpConn.SafeWrite(readyP)
		}
		c.sh.Lock.Unlock()
	}
}

func (c *OverlayClient) requestStreamsFromServer() {
	packet := getstreams.NewGetStreamsPacket(true)
	c.rpConn.SafeWrite(packet)
}

func (c *OverlayClient) requestSdpFromRP(streamName string) {
	packet := requeststream.NewRequestStreamPacket(streamName, c.overlay.LocalAddr)
	c.rpConn.SafeWrite(packet)
}

func main() {
	Neighbors, name := shell.GetRealArgs("File path with client configs", "Name of the config to read from static/configs/C${NAME}.conf")
	logger := log.New(os.Stdout, "", log.LstdFlags)
	s := shell.NewShell(logger, name)

	overlayNode := NewOverlayClient(logger, Neighbors[0], Neighbors[1:], s)
	overlayNode.registerCommands(s)
	s.Run()
}
