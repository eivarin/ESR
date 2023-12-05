package main

import (
	"io"
	"log"
	"main/ffmpeg"
	"main/overlay"
	"main/packets"
	"main/packets/getstreams"
	"main/packets/join"
	"main/packets/requeststream"
	"main/shell"
	"net"
	"os"
	"time"
)

type OverlayClient struct {
	logger   *log.Logger
	overlay  *overlay.Overlay
	rpConn   *net.TCPConn
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
	packet := packets.NewOverlayPacket(jp)
	c.rpConn, _ = net.DialTCP("tcp", nil, rpAddr)
	c.rpConn.Write(packet.Encode())
	buf := make([]byte, 1500)
	_, err := c.rpConn.Read(buf)
	if err != nil {
		log.Panic("JoinOverlayRPResponsePacket failed", err)
	}
	go func() {
		for {
			buf := make([]byte, 1500)
			_, err := c.rpConn.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Fatal("\nRP connection closed")
				} else {
					c.logger.Println("RP connection error", err)
				}
			}
			c.HandlePacketFromRP(buf, c.rpConn)
		}
	}()
}

func (c *OverlayClient) LeaveOverlay() {

}

func (c *OverlayClient) HandlePacketFromNeighbor(p *packets.OverlayPacket) {

}

func (c *OverlayClient) HandlePacketFromRP(buf []byte, conn *net.TCPConn) {
	p := packets.OverlayPacket{}
	p.Decode(buf)
	switch p.Type() {
	case packets.OverlayPacketTypeGetStreamsResponse:
		gsp, _ := p.InnerPacket().(*getstreams.GetStreamsResponsePacket)
		c.streams = make(map[string]bool)
		for _, stream := range gsp.StreamNames {
			c.streams[stream] = true
			c.logger.Printf("%s \n", stream)
		}
		c.sh.Lock.Unlock()
	case packets.OverlayPacketTypeRequestStreamResponse:
		rsp, _ := p.InnerPacket().(*requeststream.RequestStreamResponsePacket)
		if !rsp.Status {
			c.logger.Printf("Stream %s not found \n", rsp.StreamName)
		} else {
			c.ffplayer.AddInstance(rsp.StreamName, rsp.Sdp, false)
			c.logger.Printf("Playing %s \n", rsp.StreamName)
			//send confirmation to rp
		}
		c.sh.Lock.Unlock()
	}
}

func (c *OverlayClient) requestStreamsFromServer() {
	packet := getstreams.NewGetStreamsPacket(true)
	c.rpConn.Write(packets.NewOverlayPacket(packet).Encode())
}

func main() {
	Neighbors, name := shell.GetRealArgs("File path with client configs", "Name of the config to read from static/configs/C${NAME}.conf")
	logger := log.New(os.Stdout, "", log.LstdFlags)
	s := shell.NewShell(logger, name)

	overlayNode := NewOverlayClient(logger, Neighbors[0], Neighbors[1:], s)
	overlayNode.registerCommands(s)
	s.Run()
}