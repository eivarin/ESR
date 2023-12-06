package main

import (
	"main/shell"
)

func (client *OverlayClient) registerCommands(s *shell.Shell) {
	s.RegisterCommand("leave", func(args []string) {
		client.LeaveOverlay()
	})
	s.RegisterCommand("list", client.listStreams)
	s.RegisterCommand("play", client.playStream)
}

func (client *OverlayClient) listStreams(args []string) {
	client.logger.Println("Requesting streams: ")
	client.requestStreamsFromServer()
	client.sh.Lock.Lock()
}

func (client *OverlayClient) playStream(args []string) {
	if len(args) != 1 && len(args) != 2 {
		client.logger.Println("Usage: play <streamName> [debug: true/FALSE]")
		return
	}
	client.sh.Lock.Lock()
	client.requestStreamsFromServer()
	if _, ok := client.streams[args[0]]; !ok {
		client.logger.Println("Stream", args[0], "is not available in rp. Use command 'list' to update")
		return
	} else {
		client.requestSdpFromRP(args[0])
		client.sh.Lock.Lock()
	}
}