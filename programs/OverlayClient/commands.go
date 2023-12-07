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
	s.RegisterCommand("stop", client.stop)
	s.RegisterCommand("info", client.info)
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
	client.listStreams(args)
	client.sh.Lock.Lock()
	if _, ok := client.streams[args[0]]; !ok {
		client.sh.Lock.Unlock()
		client.logger.Println("Stream", args[0], "is not available in rp. Use command 'list' to update")
		return
	} else {
		client.sh.Lock.Unlock()
		client.requestSdpFromRP(args[0])
		client.sh.Lock.Lock()
	}
}

func (client *OverlayClient) stop(args []string) {
	if len(args) != 1 {
		client.logger.Println("Usage: stop <streamName>")
		return
	}
	if _, ok := client.streams[args[0]]; !ok {
		client.logger.Println("Stream", args[0], "is not available in rp. Use command 'list' to update")
		return
	} else {
		client.sh.Lock.Lock()
		client.logger.Println("Stopping stream", args[0])
		client.ffplayer.RemoveInstance(args[0])
		client.streams[args[0]] = false
		client.noLonguerInterested(args[0])
		client.sh.Lock.Unlock()
	}
}


func (client *OverlayClient) info(args []string) {
	client.logger.Println("Local address:", client.overlay.LocalAddr)
	client.logger.Println("RP address:", client.overlay.RPAddr)
	client.logger.Println("Neighbors:", client.overlay.Neighbors)
	client.logger.Println("Running streams:", client.streams)

}
