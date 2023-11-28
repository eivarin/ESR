package playground

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
)

const (
	rtpPort = 5004
)

func test() {
	sendRTPPackets()
}

// sendRTPPackets simulates sending RTP packets over UDP
func sendRTPPackets() {
	// Open the H.264 video file
	file, err := os.Open("./output_video.mp4")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	h264, err := h264reader.NewReader(file)
	if err != nil {
		fmt.Printf("Error creating H264 reader: %v", err)
		return
	}
	var nals []*h264reader.NAL
	for {
		nal, err := h264.NextNAL()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error getting next NAL: %v", err)
			return
		}
		nals = append(nals, nal)
	}
	number_of_nals := len(nals)
	fmt.Printf("Number of NALs: %d\n", number_of_nals)

	// Create a RTP packetizer
	packetizer := rtp.NewPacketizer(1400, 96, 0, &codecs.H264Payloader{}, rtp.NewRandomSequencer(), 90000)

	// Create a UDP connection to send RTP packets
	conn, err := net.Dial("udp", fmt.Sprintf("127.0.0.1:%d", rtpPort))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	var timestamp uint32
	var current_nal int
	for {
		if current_nal >= number_of_nals {
			current_nal = 0
			timestamp = 0
		}
		nal := nals[current_nal]
		current_nal += 1
		// Increment timestamp
		timestamp += 1
		// Create a new RTP packet
		packets := packetizer.Packetize(nal.Data, 3000)

		// Send the packets
		for _, p := range packets {
			bytes, _ := p.Marshal()
			_, err := conn.Write(bytes)
			if err != nil {
				conn, err = net.Dial("udp", fmt.Sprintf("127.0.0.1:%d", rtpPort))
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()
				continue
				// log.Fatal(err)
			}
			//sleep
			time.Sleep(time.Microsecond * 50)
		}
	}
}
