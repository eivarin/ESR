package main

import (
	"main/shell"
)

func (server *OverlayServer) registerCommands(s *shell.Shell) {
	s.RegisterCommand("leave", func(args []string) {
		server.LeaveOverlay()
	})
	s.RegisterCommand("list", server.listStreams)
}

func (server *OverlayServer) listStreams(args []string) {
	server.logger.Print("Listing streams:")
	for stream := range server.availableStreams {
		server.logger.Printf(" [%s:%s] ", stream, server.availableStreams[stream].Path)
	}
	server.logger.Println()
}