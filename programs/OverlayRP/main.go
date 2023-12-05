package main

import (
	"io"
	"log"
	"main/ffmpeg"
	"main/overlay"
	"main/packets"
	"main/packets/getstreams"
	"main/packets/join"
	"main/packets/publish"
	"main/packets/requeststream"
	"main/shell"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"github.com/dominikbraun/graph"
)

type Node struct {
	Addr string
	T    byte
	conn *net.TCPConn
	connLock *sync.Mutex
}

func newNode(addr string, t byte, conn *net.TCPConn) *Node {
	n := new(Node)
	n.Addr = addr
	n.T = t
	n.conn = conn
	n.connLock = &sync.Mutex{}
	return n
}

type edgeNode struct {
	Node
	rtts    time.Duration
	addr    string
	streams map[string]bool
}

func newEdgeNode(addr string, t byte, conn *net.TCPConn, rtts time.Duration, neighbor string) *edgeNode {
	n := new(edgeNode)
	n.Addr = addr
	n.T = t
	n.conn = conn
	n.rtts = rtts
	n.addr = neighbor
	n.streams = make(map[string]bool)
	return n
}

type streamRequester struct {
	node      *Node
	graphPath []string
}

type availableStream struct {
	name       string
	hosts      map[string]bool
	running    bool
	mstree     graph.Graph[string, string]
	requesters map[string]*streamRequester
	sdp 	   string
}

func (as *availableStream) addRequester(node *Node, graphPath []string) {
	as.requesters[node.Addr] = &streamRequester{node, graphPath}
}

// func (as *availableStream) removeRequester(node *Node) {
// 	delete(as.requesters, node.Addr)
// }

//new requester
//client(ask for stream) -> rp(make tree) && rp(asks for sdp) -> server(sends sdp) -> rp(sends sdp) -> client(confirms opening) -> rp(confirms opening) -> server(starts sending) -> rp(starts sending) -> client(receives)
//client(ask for stream) -> rp(remake tree) && rp(sends sdp) -> client(confirms opening) -> rp(confirms opening) -> server(starts sending) -> rp(starts sending) -> client(receives)

//rp(asks for sdp) done
//server(sends sdp) next

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
	sh 			 	 *shell.Shell
	redirecter 	 	 *ffmpeg.FFRedirecter
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
	rp.fullTree = graph.New(graph.StringHash)
	rp.overlayTree = graph.New(graph.StringHash)
	rp.fullTree.AddVertex(rpAddr)
	rp.overlayTree.AddVertex(rpAddr)
	rp.availableStreams = make(map[string]*availableStream)
	addr, _ := net.ResolveTCPAddr("tcp", rpAddr+":5000")
	rp.rpListener, _ = net.ListenTCP("tcp", addr)
	go func() {
		for {
			conn, _ := rp.rpListener.AcceptTCP()
			go func(conn *net.TCPConn) {
				buf := make([]byte, 1500)
				conn.Read(buf)
				rp.addNode(buf, conn)
			}(conn)
		}
	}()
	rp.dotLock = &sync.Mutex{}
	return rp
}

func (rp *OverlayRP) addNode(buf []byte, conn *net.TCPConn) {
	p := packets.OverlayPacket{}
	p.Decode(buf)
	jp := p.InnerPacket().(*join.JoinOverlayRPPacket)
	addr := jp.LocalAddr
	rp.insertNode(addr, jp.OverlayType, conn, jp.NeighborsRTTs[jp.NeighborsAddrs[0]], jp.NeighborsAddrs[0])
	resp := packets.NewOverlayPacket(join.NewJoinOverlayRPResponsePacket(true))
	conn.Write(resp.Encode())
	go func(con *net.TCPConn) {
		for {
			buf := make([]byte, 1500)
			_, err := con.Read(buf)
			if err != nil {
				if err == io.EOF {
					rp.handleDisconectedNode(addr, jp.OverlayType)
				} else {
					rp.logger.Println("Error reading from", addr, err)
				}
				return
			}
			rp.HandlePacketFromRP(buf, con)
		}
	}(conn)
	// log added node with neighbors
	rp.logger.Println("Added node", addr, "with neighbors", jp.NeighborsAddrs)
	rp.nodesLock.RLock()
	rp.drawGraph(rp.fullTree)
	rp.nodesLock.RUnlock()
}

func (rp *OverlayRP) insertNode(addr string, t byte, conn *net.TCPConn, rtts time.Duration, neighbor string) {
	rp.nodesLock.Lock()
	switch t {
	case overlay.NodeServer:
		rp.servers[addr] = newEdgeNode(addr, overlay.NodeServer, conn, rtts, neighbor)
	case overlay.NodeClient:
		rp.clients[addr] = newEdgeNode(addr, overlay.NodeClient, conn, rtts, neighbor)
	case overlay.NodeOverlay:
		rp.nodes[addr] = newNode(addr, overlay.NodeOverlay, conn)
	}
	rp.addVertexToGraph(rp.fullTree, addr, []string{neighbor}, map[string]time.Duration{neighbor: rtts})
	if t == overlay.NodeOverlay {
		rp.addVertexToGraph(rp.overlayTree, addr, []string{neighbor}, map[string]time.Duration{neighbor: rtts})
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
		delete(rp.clients, addr)
		rp.removeVertexFromGraph(rp.fullTree, addr)
		//drop client from requesting streams
		//drop empty streams
	case overlay.NodeOverlay:
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

func (rp *OverlayRP) HandlePacketFromNeighbor(p *packets.OverlayPacket) {

}

func (rp *OverlayRP) HandlePacketFromRP(buf []byte, conn *net.TCPConn) {
	p := packets.OverlayPacket{}
	p.Decode(buf)
	switch p.T {
	case packets.OverlayPacketTypePublish:
		publishPacket := p.InnerPacket().(*publish.PublishPacket)
		for _, streamName := range publishPacket.StreamNames {
			if _, ok := rp.availableStreams[streamName]; !ok {
				rp.availableStreams[streamName] = &availableStream{streamName, make(map[string]bool), false, nil, make(map[string]*streamRequester), ""}
			} 
			rp.availableStreams[streamName].hosts[publishPacket.StreamerAddr] = true
			rp.nodesLock.Lock()
			rp.servers[publishPacket.StreamerAddr].streams[streamName] = true
			rp.nodesLock.Unlock()
		}
		publishPacketResponse := publish.NewPublishResponsePacket(len(publishPacket.StreamNames))
		rp.nodesLock.RLock()
		rp.servers[publishPacket.StreamerAddr].conn.Write(packets.NewOverlayPacket(publishPacketResponse).Encode())
		rp.nodesLock.RUnlock()
	case packets.OverlayPacketTypeGetStreams:
		streamNames := make([]string, 0)
		for streamName := range rp.availableStreams {
			streamNames = append(streamNames, streamName)
		}
		getStreamsPacketResponse := getstreams.NewGetStreamsResponsePacket(streamNames)
		conn.Write(packets.NewOverlayPacket(getStreamsPacketResponse).Encode())
	case packets.OverlayPacketTypeRequestStream:
		requestPacket := p.InnerPacket().(*requeststream.RequestStreamPacket)
		if as, ok := rp.availableStreams[requestPacket.StreamName]; !ok {
			failure := requeststream.NewRequestStreamResponsePacket(false, requestPacket.StreamName, "",)
			conn.Write(packets.NewOverlayPacket(failure).Encode())
		} else {
			if as.running {
				//send sdp
				success := requeststream.NewRequestStreamResponsePacket(true, requestPacket.StreamName, as.sdp)
				conn.Write(packets.NewOverlayPacket(success).Encode())
			} else {
				//get sdp to send
				requestSDP := requeststream.NewRequestStreamPacket(requestPacket.StreamName, "")
				rp.nodesLock.Lock()
				defer rp.nodesLock.Unlock()
				as.running = true
				as.addRequester(&rp.clients[requestPacket.RequesterAddr].Node, nil)
				rp.getBestServerForStream(requestPacket.StreamName).Write(packets.NewOverlayPacket(requestSDP).Encode())
			}
		}
	}
}

//fix this
func (rp *OverlayRP) getBestServerForStream(streamName string) *net.TCPConn {
	rp.nodesLock.RLock()
	defer rp.nodesLock.RUnlock()
	bestServer := ""
	minRTT := time.Duration(math.MaxInt64-1)
	for serverAddr, server := range rp.servers {
		if server.rtts < minRTT && server.streams[streamName] {
			bestServer = serverAddr
			minRTT = server.rtts
		}
	}
	return rp.servers[bestServer].conn
}

func main() {
	Neighbors, name := shell.GetRealArgs("File with RP Addr", "Name of the node to read from statics/configs/NAME.conf")
	logger := log.New(os.Stdout, "", log.LstdFlags)
	s := shell.NewShell(logger, name)
	overlayRP := NewOverlayRP(logger, Neighbors[0], s)
	overlayRP.registerCommands(s)
	s.Run()
}
