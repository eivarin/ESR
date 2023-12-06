package main

import (
	"fmt"
	"io"
	"log"
	"main/ffmpeg"
	"main/overlay"
	"main/packets"
	"main/packets/getstreams"
	"main/packets/join"
	"main/packets/publish"
	"main/packets/ready"
	"main/packets/redirects"
	"main/packets/requeststream"
	"main/safesockets"
	"main/shell"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"github.com/dominikbraun/graph"
)

type Node struct {
	Addr     string
	T        byte
	conn    *safesockets.SafeTCPConn
}

func newNode(addr string, t byte, conn *safesockets.SafeTCPConn) *Node {
	n := new(Node)
	n.Addr = addr
	n.T = t
	n.conn = conn
	return n
}

type edgeNode struct {
	Node
	rtts     time.Duration
	neighbor string
	streams  map[string]bool
}

func newEdgeNode(addr string, t byte, conn *safesockets.SafeTCPConn, rtts time.Duration, neighbor string) *edgeNode {
	en := new(edgeNode)
	en.Addr = addr
	en.T = t
	en.conn = conn
	en.rtts = rtts
	en.neighbor = neighbor
	en.streams = make(map[string]bool)
	return en
}

type streamRequester struct {
	node      *edgeNode
	graphPath []string
}

type availableStream struct {
	name       string
	hosts      map[string]bool
	pickedHost string
	running    bool
	mstree     graph.Graph[string, string]
	mstreelock *sync.RWMutex
	requesters map[string]*streamRequester
	sdp        string
	sdpLock   *sync.RWMutex
}

func newAvailableStream(name string) *availableStream {
	as := new(availableStream)
	as.name = name
	as.hosts = make(map[string]bool)
	as.running = false
	as.mstree = graph.New(graph.StringHash)
	as.mstreelock = &sync.RWMutex{}
	as.requesters = make(map[string]*streamRequester)
	as.sdp = ""
	as.sdpLock = &sync.RWMutex{}
	return as
}

func (as *availableStream) addRequester(n *edgeNode, graphPath []string) {
	as.requesters[n.Addr] = &streamRequester{n, graphPath}
}

type OverlayRP struct {
	logger           *log.Logger
	overlay          *overlay.Overlay
	rpListener       *net.TCPListener
	nodesLock        *sync.RWMutex
	nodes            map[string]*Node
	servers          map[string]*edgeNode
	clients          map[string]*edgeNode
	availableStreams map[string]*availableStream
	fullTree         graph.Graph[string, string]
	overlayTree      graph.Graph[string, string]
	dotLock          *sync.Mutex
	sh               *shell.Shell
	redirecter       *ffmpeg.FFRedirecter
}

func NewOverlayRP(logger *log.Logger, rpAddr string, sh *shell.Shell) *OverlayRP {
	rp := new(OverlayRP)
	rp.sh = sh
	rp.logger = logger
	rp.overlay = overlay.NewOverlayRP(logger, 5000, rpAddr, rp)
	rp.nodes = make(map[string]*Node)
	rp.servers = make(map[string]*edgeNode)
	rp.clients = make(map[string]*edgeNode)
	udpAddr, _ := net.ResolveUDPAddr("udp", ":4999")
	rp.redirecter = ffmpeg.NewFFRedirecter(logger, udpAddr, false)
	rp.nodesLock = &sync.RWMutex{}
	rp.nodes[rpAddr] = newNode(rpAddr, overlay.NodeOverlay, nil)
	rp.fullTree = graph.New[string, string](graph.StringHash, graph.Weighted())
	rp.overlayTree = graph.New[string, string](graph.StringHash, graph.Weighted())
	rp.fullTree.AddVertex(rpAddr)
	rp.overlayTree.AddVertex(rpAddr)
	rp.availableStreams = make(map[string]*availableStream)
	addr, _ := net.ResolveTCPAddr("tcp", rpAddr+":5000")
	rp.rpListener, _ = net.ListenTCP("tcp", addr)
	go func() {
		for {
			conn, _ := rp.rpListener.AcceptTCP()
			go func(conn *net.TCPConn) {
				safeConn := safesockets.NewSafeTCPConn(conn)
				rp.addNode(safeConn)
			}(conn)
		}
	}()
	rp.dotLock = &sync.Mutex{}
	return rp
}

func (rp *OverlayRP) addNode(conn *safesockets.SafeTCPConn) {
	p, _, _, _ := conn.SafeRead()
	jp := p.(*join.JoinOverlayRPPacket)
	addr := jp.LocalAddr
	if jp.OverlayType == overlay.NodeOverlay {
		rp.insertOverlayNode(addr, jp.OverlayType, conn, jp.NeighborsRTTs, jp.NeighborsAddrs)
	} else {
		rp.insertEdgeNode(addr, jp.OverlayType, conn, jp.NeighborsRTTs[jp.NeighborsAddrs[0]], jp.NeighborsAddrs[0])
	}
	resp := join.NewJoinOverlayRPResponsePacket(true)
	conn.SafeWrite(resp)
	go func(con *safesockets.SafeTCPConn) {
		for {
			p, t, _, err := con.SafeRead()
			if err != nil {
				if err == io.EOF {
					rp.handleDisconectedNode(addr, jp.OverlayType)
				} else {
					rp.logger.Println("Error reading from", addr, err)
				}
				return
			}
			rp.HandlePacketFromRP(p, t, con)
		}
	}(conn)
	// log added node with neighbors
	switch jp.OverlayType {
	case overlay.NodeServer:
		rp.logger.Println("Server", addr, "connected with neighbor", jp.NeighborsAddrs[0])
	case overlay.NodeClient:
		rp.logger.Println("Client", addr, "connected with neighbor", jp.NeighborsAddrs[0])
	case overlay.NodeOverlay:
		rp.logger.Println("Node", addr, "connected with neighbors", jp.NeighborsAddrs)
	}
	rp.nodesLock.RLock()
	rp.drawGraph(rp.fullTree)
	rp.nodesLock.RUnlock()
}

func (rp *OverlayRP) insertOverlayNode(name string, t byte, conn *safesockets.SafeTCPConn, rtts map[string]time.Duration, neighbors []string) {
	rp.nodesLock.Lock()
	rp.nodes[name] = newNode(name, t, conn)
	rp.addVertexToGraph(rp.fullTree, name, neighbors, rtts)
	rp.addVertexToGraph(rp.overlayTree, name, neighbors, rtts)
	rp.nodesLock.Unlock()
}

func (rp *OverlayRP) insertEdgeNode(name string, t byte, conn *safesockets.SafeTCPConn, rtts time.Duration, neighbor string) {
	rp.nodesLock.Lock()
	switch t {
	case overlay.NodeServer:
		rp.servers[name] = newEdgeNode(name, t, conn, rtts, neighbor)
		rp.addVertexToGraph(rp.fullTree, name, []string{neighbor}, map[string]time.Duration{neighbor: rtts})
	case overlay.NodeClient:
		rp.clients[name] = newEdgeNode(name, t, conn, rtts, neighbor)
		rp.addVertexToGraph(rp.fullTree, name, []string{neighbor}, map[string]time.Duration{neighbor: rtts})
	}
	rp.nodesLock.Unlock()
}

func (rp *OverlayRP) handleDisconectedNode(addr string, t byte) {
	rp.nodesLock.Lock()
	switch t {
	case overlay.NodeServer:
		rp.logger.Println("Server", addr, "disconnected")
		for streamName := range rp.servers[addr].streams {
			delete(rp.availableStreams[streamName].hosts, addr)
		}
		delete(rp.servers, addr)
		rp.removeVertexFromGraph(rp.fullTree, addr)
		//handle dropped streams
	case overlay.NodeClient:
		rp.logger.Println("Client", addr, "disconnected")
		delete(rp.clients, addr)
		rp.removeVertexFromGraph(rp.fullTree, addr)
		//drop client from requesting streams
		//drop empty streams
	case overlay.NodeOverlay:
		rp.logger.Println("Node", addr, "disconnected")
		delete(rp.nodes, addr)
		rp.removeVertexFromGraph(rp.fullTree, addr)
		rp.removeVertexFromGraph(rp.overlayTree, addr)
		//recalc all streams paths
	}
	rp.nodesLock.Unlock()
}

func (rp *OverlayRP) JoinOverlay(rpAddr *net.TCPAddr, localAddr string, neighborsAddrs []string, neighborsRTTs map[string]time.Duration) {
}

func (rp *OverlayRP) LeaveOverlay() {

}

func (rp *OverlayRP) HandlePacketFromNeighbor(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn) {

}

func (rp *OverlayRP) HandlePacketFromRP(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn) {
	switch t {
	case packets.OverlayPacketTypePublish:
		publishPacket := p.(*publish.PublishPacket)
		for _, streamName := range publishPacket.StreamNames {
			if _, ok := rp.availableStreams[streamName]; !ok {
				rp.availableStreams[streamName] = newAvailableStream(streamName)
			}
			rp.availableStreams[streamName].hosts[publishPacket.StreamerAddr] = true
			rp.nodesLock.Lock()
			rp.servers[publishPacket.StreamerAddr].streams[streamName] = true
			rp.nodesLock.Unlock()
		}
		publishPacketResponse := publish.NewPublishResponsePacket(len(publishPacket.StreamNames))
		rp.writeToServer(publishPacket.StreamerAddr, publishPacketResponse)
	case packets.OverlayPacketTypeGetStreams:
		streamNames := make([]string, 0)
		for streamName := range rp.availableStreams {
			streamNames = append(streamNames, streamName)
		}
		getStreamsPacketResponse := getstreams.NewGetStreamsResponsePacket(streamNames)
		conn.SafeWrite(getStreamsPacketResponse)
	case packets.OverlayPacketTypeRequestStream:
		requestPacket := p.(*requeststream.RequestStreamPacket)
		if as, ok := rp.availableStreams[requestPacket.StreamName]; !ok {
			failure := requeststream.NewRequestStreamResponsePacket(false, requestPacket.StreamName, "")
			conn.SafeWrite(failure)
		} else {
			as.sdpLock.RLock()
			if as.running{
				success := requeststream.NewRequestStreamResponsePacket(true, requestPacket.StreamName, as.sdp)
				as.addRequester(rp.clients[requestPacket.RequesterAddr], nil)
				rp.generatePathsForStream(as)
				rp.logger.Println("Sending redirects")
				redirectsMap := make(map[string][]string)
				for _, requester := range as.requesters {
					for i := 1; i < len(requester.graphPath)-1; i++ {
						if _, ok := redirectsMap[requester.graphPath[i]]; !ok {
							redirectsMap[requester.graphPath[i]] = make([]string, 0)
						}
						redirectsMap[requester.graphPath[i]] = append(redirectsMap[requester.graphPath[i]], requester.graphPath[i+1])
					}
				}
				for nodeAddr, reds := range redirectsMap {
					rp.writeToNode(nodeAddr, redirects.NewRedirectsPacket(as.name, reds))
				}
				conn.SafeWrite(success)
				as.sdpLock.RUnlock()
			} else {
				as.sdpLock.RUnlock()
				requestSDP := requeststream.NewRequestStreamPacket(requestPacket.StreamName, "")
				as.sdpLock.Lock()
				as.running = true
				as.addRequester(rp.clients[requestPacket.RequesterAddr], nil)
				rp.getBestServerForStream(as).SafeWrite(requestSDP)
				rp.generatePathsForStream(as)
			}
		}
	case packets.OverlayPacketTypeRequestStreamResponse:
		requestPacketResponse := p.(*requeststream.RequestStreamResponsePacket)
		if !requestPacketResponse.Status {
			rp.logger.Println("Stream", requestPacketResponse)
			rp.logger.Println("Stream", requestPacketResponse.StreamName, "not found by Server")
		} else if as, ok := rp.availableStreams[requestPacketResponse.StreamName]; ok {
			as.sdp = requestPacketResponse.Sdp
			as.sdpLock.Unlock()
			as.mstreelock.RLock()
			rp.configureFFRedirecterWithStream(as)
			redirectsMap := make(map[string][]string)
			for _, requester := range as.requesters {
				for i := 1; i < len(requester.graphPath)-1; i++ {
					if _, ok := redirectsMap[requester.graphPath[i]]; !ok {
						redirectsMap[requester.graphPath[i]] = make([]string, 0)
					}
					redirectsMap[requester.graphPath[i]] = append(redirectsMap[requester.graphPath[i]], requester.graphPath[i+1])
				}
			}
			for nodeAddr, reds := range redirectsMap {
				rp.writeToNode(nodeAddr, redirects.NewRedirectsPacket(as.name, reds))
			}
			for _, requester := range as.requesters {
				requester.node.conn.SafeWrite(requestPacketResponse)
			}
			as.mstreelock.RUnlock()
		} else {
			rp.logger.Println("Stream", requestPacketResponse.StreamName, "not found by RP")
		}
	case packets.OverlayPacketTypeReady:
		readyPacket := p.(*ready.ReadyPacket)
		if as, ok := rp.availableStreams[readyPacket.StreamName]; ok {
			as.mstreelock.RLock()
			rp.writeToServer(as.pickedHost, readyPacket)
			as.mstreelock.RUnlock()
		}
	}
}

func (rp *OverlayRP) configureFFRedirecterWithStream(as *availableStream) {
	destinations := make(map[string]*safesockets.SafeUDPWritter)
	for _, requester := range rp.availableStreams[as.name].requesters {
		if _, ok := destinations[requester.node.Addr]; !ok {
			destAddr, _ := net.ResolveUDPAddr("udp", requester.graphPath[1]+":4999")
			conn, _ := net.DialUDP("udp", nil, destAddr)
			destinations[requester.node.Addr] = safesockets.NewSafeUDPWritter(conn)
		}
	}
	rp.redirecter.AddStream(as.name, destinations)
}

// fix this
func (rp *OverlayRP) getBestServerForStream(as *availableStream) *safesockets.SafeTCPConn {
	rp.nodesLock.RLock()
	defer rp.nodesLock.RUnlock()
	bestServer := ""
	minRTT := time.Duration(math.MaxInt64 - 1)
	for serverAddr, server := range rp.servers {
		if server.rtts < minRTT && server.streams[as.name] {
			bestServer = serverAddr
			minRTT = server.rtts
		}
	}
	as.pickedHost = bestServer
	return rp.servers[bestServer].conn
}

func (rp *OverlayRP) writeToServer(serverAddr string, p packets.PacketI) {
	rp.nodesLock.RLock()
	defer rp.nodesLock.RUnlock()
	if server, ok := rp.servers[serverAddr]; ok {
		server.conn.SafeWrite(p)
	}
}

// func (rp *OverlayRP) writeToClient(clientAddr string, p packets.PacketI) {
// 	rp.nodesLock.RLock()
// 	defer rp.nodesLock.RUnlock()
// 	if client, ok := rp.clients[clientAddr]; ok {
// 		client.conn.SafeWrite(p)
// 	}
// }

func (rp *OverlayRP) writeToNode(nodeAddr string, p packets.PacketI) {
	rp.nodesLock.RLock()
	defer rp.nodesLock.RUnlock()
	fmt.Println(nodeAddr, p)
	if node, ok := rp.nodes[nodeAddr]; ok {
		node.conn.SafeWrite(p)
	}
}

func main() {
	Neighbors, name := shell.GetRealArgs("File with RP Addr", "Name of the node to read from statics/configs/NAME.conf")
	logger := log.New(os.Stdout, "", log.LstdFlags)
	s := shell.NewShell(logger, name)
	overlayRP := NewOverlayRP(logger, Neighbors[0], s)
	overlayRP.registerCommands(s)
	s.Run()
}
