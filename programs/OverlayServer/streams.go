package main

import (
	"main/packets"
	"main/packets/publish"
	"os"
	"strings"
)

type availableStream struct {
	Path string
	Running bool
}

func NewAvailableStream(path string) *availableStream {
	s := availableStream{}
	s.Path = path
	s.Running = false
	return &s
}

func (server *OverlayServer) populateFromManifest(manifestPath string) {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		server.logger.Println("Error reading file:", err)
		os.Exit(1)
	}

	// Split the content into lines and populate the Neighbors slice
	Lines := strings.Split(string(content), "\n")

	for _, line := range Lines {
		if line != "" {
			// Split the line into words and populate the Neighbors slice
			Words := strings.Split(line, " ")
			server.logger.Println("Line:", Words[1])
			server.availableStreams[Words[0]] = NewAvailableStream(Words[1])
		}
	}
}

func (server *OverlayServer) publishStreams() {
	streamNames := make([]string, len(server.availableStreams))
	i := 0
	for k := range server.availableStreams {
		streamNames[i] = k
		i++
	}

	publishPacket := publish.NewPublishPacket(server.localAddr, streamNames)
	packet := packets.NewOverlayPacket(publishPacket)
	server.rpConn.Write(packet.Encode())
	buf := make([]byte, 1500)
	server.rpConn.Read(buf)
	rPacket := packets.OverlayPacket{}
	rPacket.Decode(buf)
	rPublish := rPacket.InnerPacket().(*publish.PublishResponsePacket)
	if rPublish.NumberOfStreamsOk != len(streamNames) {
		server.logger.Println("Publish failed")
		panic("Publish failed")
	}
}