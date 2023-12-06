package main

import (
	"main/shell"
)

func (rp *OverlayRP) registerCommands(s *shell.Shell) {
	s.RegisterCommand("leave", func(args []string) {
		rp.LeaveOverlay()
	})
	s.RegisterCommand("listStreams", rp.listStreams)
	s.RegisterCommand("listServers", rp.listServers)
	s.RegisterCommand("listClients", rp.listClients)
	s.RegisterCommand("listNodes", rp.listNodes)
	s.RegisterCommand("listFullTree", rp.listFullTree)
	s.RegisterCommand("listOverlayTree", rp.listOverlayTree)
	s.RegisterCommand("info", rp.info)
}

func (rp *OverlayRP) listStreams(args []string) {
	rp.logger.Println("Listing streams:")
	for _, stream := range rp.availableStreams {
		hosts := []string{}
		for host := range stream.hosts {
			hosts = append(hosts, host)
		}
		rp.logger.Printf("Stream %s: %v\n", stream.name, hosts)
	}
}

func (rp *OverlayRP) listServers(args []string) {
	rp.logger.Println("Listing servers:")
	for _, server := range rp.servers {
		rp.logger.Printf("Server %s: %v\n", server.neighbor, server.streams)
	}
}

func (rp *OverlayRP) listClients(args []string) {
	rp.logger.Println("Listing clients:")
	for _, client := range rp.clients {
		rp.logger.Printf("Client %s: %v\n", client.neighbor, client.streams)
	}
}

func (rp *OverlayRP) listNodes(args []string) {
	rp.logger.Println("Listing nodes:")
	for _, node := range rp.nodes {
		rp.logger.Printf("Node %s\n", node.Addr)
	}
}

func (rp *OverlayRP) listFullTree(args []string) {
	rp.logger.Println("Listing full tree:")
	edges, _ := rp.fullTree.Edges()
	for _, edge := range edges {
		rp.logger.Printf("%s -> %s\n", edge.Source, edge.Target)
	}
}

func (rp *OverlayRP) listOverlayTree(args []string) {
	rp.logger.Println("Listing overlay tree:")
	edges, _ := rp.overlayTree.Edges()
	for _, edge := range edges {
		rp.logger.Printf("%s -> %s\n", edge.Source, edge.Target)
	}
}

func (rp *OverlayRP) info(args []string) {
	rp.logger.Println("Info:")
	rp.logger.Printf("Nodes: %d\n", len(rp.nodes))
	rp.listNodes(args)
	rp.logger.Printf("Servers: %d\n", len(rp.servers))
	rp.listServers(args)
	rp.logger.Printf("Clients: %d\n", len(rp.clients))
	rp.listClients(args)
	rp.logger.Printf("Available streams: %d\n", len(rp.availableStreams))
	rp.listStreams(args)
}
