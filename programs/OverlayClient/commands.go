package main

import (
	"main/shell"
)

func (client *OverlayClient) registerCommands(s *shell.Shell) {
	s.RegisterCommand("leave", func(args []string) {
		client.LeaveOverlay()
	})
	s.RegisterCommand("listStreams", client.listStreams)
}

func (client *OverlayClient) listStreams(args []string) {
	client.logger.Println("Requesting streams: ")
	client.requestStreamsFromServer()
	client.sh.Lock.Lock()
}