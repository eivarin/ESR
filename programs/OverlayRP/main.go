package main

import (
	"io"
	"log"
	"main/ffmpeg"
	"main/overlay"
	"main/packets"
	"main/packets/getstreams"
	"main/packets/join"
	"main/packets/nolonguer"
	"main/packets/publish"
	"main/packets/ready"
	"main/packets/redirects"
	"main/packets/requeststream"
	"main/packets/rtt"
	"main/packets/startstream"
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
	Addr    string
	T       byte
	conn    *safesockets.SafeTCPConn
	streams map[string]bool
	lock    *sync.RWMutex
}

func newNode(addr string, t byte, conn *safesockets.SafeTCPConn) *Node {
	n := new(Node)
	n.Addr = addr
	n.T = t
	n.conn = conn
	n.streams = make(map[string]bool)
	n.lock = &sync.RWMutex{}
	return n
}

type edgeNode struct {
	Node
	rtts     time.Duration
	neighbor string
	rpRtt    time.Duration
	rttLock  *sync.RWMutex
}

func newEdgeNode(addr string, t byte, conn *safesockets.SafeTCPConn, rtts time.Duration, neighbor string) *edgeNode {
	en := new(edgeNode)
	en.Addr = addr
	en.T = t
	en.conn = conn
	en.rtts = rtts
	en.neighbor = neighbor
	en.streams = make(map[string]bool)
	en.lock = &sync.RWMutex{}
	en.rttLock = &sync.RWMutex{}
	return en
}

type streamRequester struct {
	node      *edgeNode
	graphPath []string
}

type availableStream struct {
	name         string
	hosts        map[string]bool
	pickedHost   string
	running      bool
	mstree       graph.Graph[string, string]
	mstreelock   *sync.RWMutex
	requesters   map[string]*streamRequester
	reqsReadyNum int
	startTime    time.Time
	sdp          string
	sdpLock      *sync.RWMutex
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
	as.startTime = time.Unix(0, 0)
	as.reqsReadyNum = 0
	return as
}

func (as *availableStream) addRequester(n *edgeNode, graphPath []string) {
	n.streams[as.name] = true
	as.requesters[n.Addr] = &streamRequester{n, graphPath}
}

type OverlayRP struct {
	logger           *log.Logger
	overlay          *overlay.Overlay
	rpListener       *net.TCPListener
	mainLock         *sync.RWMutex
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
	rp.mainLock = &sync.RWMutex{}
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
	rp.mainLock.RLock()
	rp.drawGraph(rp.fullTree)
	rp.mainLock.RUnlock()
}

func (rp *OverlayRP) insertOverlayNode(name string, t byte, conn *safesockets.SafeTCPConn, rtts map[string]time.Duration, neighbors []string) {
	rp.mainLock.Lock()
	rp.nodes[name] = newNode(name, t, conn)
	rp.addVertexToGraph(rp.fullTree, name, neighbors, rtts)
	rp.addVertexToGraph(rp.overlayTree, name, neighbors, rtts)
	rp.mainLock.Unlock()
}

func (rp *OverlayRP) insertEdgeNode(name string, t byte, conn *safesockets.SafeTCPConn, rtts time.Duration, neighbor string) {
	rp.mainLock.Lock()
	switch t {
	case overlay.NodeServer:
		rp.servers[name] = newEdgeNode(name, t, conn, rtts, neighbor)
		rp.addVertexToGraph(rp.fullTree, name, []string{neighbor}, map[string]time.Duration{neighbor: rtts})
	case overlay.NodeClient:
		rp.clients[name] = newEdgeNode(name, t, conn, rtts, neighbor)
		rp.addVertexToGraph(rp.fullTree, name, []string{neighbor}, map[string]time.Duration{neighbor: rtts})
	}
	rp.mainLock.Unlock()
}

func (rp *OverlayRP) handleDisconectedNode(addr string, t byte) {
	switch t {
	case overlay.NodeServer:
		rp.removeServer(addr)
	case overlay.NodeClient:
		rp.removeClient(addr)
	case overlay.NodeOverlay:
		rp.removeNode(addr)
	}
}

func (rp *OverlayRP) removeNode(addr string) {
	rp.mainLock.Lock()
	rp.logger.Println("Node", addr, "disconnected")
	streamsToChange := make([]string, 0)
	rp.logger.Println("Removing node", addr, "from streams", rp.nodes[addr].streams)
	for streamName := range rp.nodes[addr].streams {
		streamsToChange = append(streamsToChange, streamName)
	}
	delete(rp.nodes, addr)
	rp.removeVertexFromGraph(rp.fullTree, addr)
	rp.removeVertexFromGraph(rp.overlayTree, addr)
	rp.mainLock.Unlock()
	for _, streamName := range streamsToChange {
		rp.logger.Println("Stream", streamName, "needs to be reconfigured")
		if as, ok := rp.availableStreams[streamName]; ok {
			rp.reconfigureOverlayWithRedirects(as)
		}
	}
}

func (rp *OverlayRP) removeServer(addr string) {
	rp.mainLock.Lock()
	rp.logger.Println("Server", addr, "disconnected")
	for streamName := range rp.servers[addr].streams {
		rp.logger.Print("Trying to revive stream", streamName, "from server", addr)
		if as, ok := rp.availableStreams[streamName]; ok {
			delete(rp.availableStreams[streamName].hosts, addr)
			if len(rp.availableStreams[streamName].hosts) == 0 {
				nolonguerPacket := nolonguer.NewNoLonguerAvailablePacket(streamName)
				for _, requester := range as.requesters {
					requester.node.conn.SafeWrite(nolonguerPacket)
				}
				delete(rp.availableStreams, streamName)
			} else if rp.availableStreams[streamName].pickedHost == addr {
				conn := rp.getBestServerForStream(as)
				secondsSincewStartInt64 := int64(time.Since(as.startTime).Seconds()) + 5
				ssp := startstream.NewStartStreamPacket(as.name, secondsSincewStartInt64)
				conn.SafeWrite(ssp)
			}
		} else {
			rp.logger.Println("Tried to remove stream", streamName, "from server", addr, "but it was not found by RP")
		}
	}
	delete(rp.servers, addr)
	rp.removeVertexFromGraph(rp.fullTree, addr)
	rp.mainLock.Unlock()
}

func (rp *OverlayRP) removeClient(addr string) {
	rp.mainLock.RLock()
	rp.logger.Println("Client", addr, "disconnected")
	streamsToCheck := make(map[string]bool)
	for streamName := range rp.clients[addr].streams {
		rp.logger.Println("Trying to stop stream", streamName, "from client", addr)
		streamsToCheck[streamName] = true
	}
	for streamName := range streamsToCheck {
		if as, ok := rp.availableStreams[streamName]; ok {
			delete(as.requesters, addr)
			if checkIfStreamCanStop(as) {
				rp.stopStream(as)
			} else {
				rp.reconfigureOverlayWithRedirects(as)
			}
		} else {
			rp.logger.Println("Tried to remove stream", streamName, "from client", addr, "but it was not found by RP")
		}
	}
	rp.mainLock.RUnlock()
	rp.mainLock.Lock()
	delete(rp.clients, addr)
	rp.removeVertexFromGraph(rp.fullTree, addr)
	rp.mainLock.Unlock()
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
			rp.mainLock.RLock()
			server := rp.servers[publishPacket.StreamerAddr]
			rp.mainLock.RUnlock()
			server.lock.Lock()
			server.streams[streamName] = true
			server.lock.Unlock()
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
			if as.running {
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
			rp.configureOverlayWithRedirects(as)
			for _, requester := range as.requesters {
				requester.node.conn.SafeWrite(requestPacketResponse)
			}
		} else {
			rp.logger.Println("Stream", requestPacketResponse.StreamName, "not found by RP")
		}
	case packets.OverlayPacketTypeReady:
		readyPacket := p.(*ready.ReadyPacket)
		if as, ok := rp.availableStreams[readyPacket.StreamName]; ok {
			as.reqsReadyNum++
			if as.startTime == time.Unix(0, 0) {
				as.startTime = time.Now()
				ssp := startstream.NewStartStreamPacket(readyPacket.StreamName, 0)
				rp.writeToServer(as.pickedHost, ssp)
			}
		}
	case packets.OverlayPacketTypeRttResponse:
		rttPacketResp := p.(*rtt.RttResponsePacket)
		rad, _ := net.ResolveUDPAddr("udp", rttPacketResp.PeerAddr+":4994")
		rConn, _ := net.DialUDP("udp", nil, rad)
		lad, _ := net.ResolveUDPAddr("udp", rp.overlay.RPAddr+":4994")
		lConn, _ := net.ListenUDP("udp", lad)
		rp.servers[rttPacketResp.PeerAddr].rpRtt = rp.measureRTTFromRP(lConn, rConn, rttPacketResp.PeerAddr, rp.overlay.RPAddr)
		rp.servers[rttPacketResp.PeerAddr].rttLock.Unlock()

	case packets.OverlayPacketTypeNoLonguerInterested:
		nli := p.(*nolonguer.NoLonguerInterestedPacket)
		rp.logger.Println("Client", nli.ClientAddress,"no longuer interested in stream", nli.StreamName)
		if as, ok := rp.availableStreams[nli.StreamName]; ok {
			delete(as.requesters, nli.ClientAddress)
			if checkIfStreamCanStop(as) {
				rp.stopStream(as)
			} else {
				rp.reconfigureOverlayWithRedirects(as)
			}
		} else {
			rp.logger.Println("Tried to remove stream", nli.StreamName, "from client", nli.ClientAddress, "but it was not found by RP")
		}
	}
}

func checkIfStreamCanStop(as *availableStream) bool {
	return len(as.requesters) == 0
}

func (rp *OverlayRP) stopStream(as *availableStream) {
	as.running = false
	rp.writeToServer(as.pickedHost, nolonguer.NewNoLonguerInterestedPacket(as.name, ""))
	as.pickedHost = ""
}

func (rp *OverlayRP) reconfigureOverlayWithRedirects(as *availableStream) {
	nodesToClean := make(map[string]bool)
	for _, requester := range as.requesters {
		for _, pathPart := range requester.graphPath[1 : len(requester.graphPath)-1] {
			if _, ok := rp.nodes[pathPart]; ok {
				delete(rp.nodes[pathPart].streams, as.name)
			}
			nodesToClean[pathPart] = true
		}
	}
	emptyList := make([]string, 0)
	redsPacket := redirects.NewRedirectsPacket(as.name, emptyList)
	for nodeAddr := range nodesToClean {
		rp.writeToNode(nodeAddr, redsPacket)
	}
	rp.configureOverlayWithRedirects(as)
}

func (rp *OverlayRP) configureOverlayWithRedirects(as *availableStream) {
	rp.generatePathsForStream(as)
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
		redsPacket := redirects.NewRedirectsPacket(as.name, reds)
		n := rp.nodes[nodeAddr]
		n.lock.Lock()
		n.streams[as.name] = true
		n.lock.Unlock()
		rp.logger.Println("Sending redirects to node", nodeAddr, "for stream", as.name, "with reds", reds)
		rp.writeToNode(nodeAddr, redsPacket)
	}
	as.mstreelock.RUnlock()
}

func (rp *OverlayRP) measureRTTFromRP(listenningConn net.Conn, writtingConn net.Conn, addr string, localAddr string) time.Duration {
	defer writtingConn.Close()
	defer listenningConn.Close()

	buf := make([]byte, 10)
	now := time.Now()

	writtingConn.Write([]byte("rtt"))
	listenningConn.Read(buf)
	rtt := time.Since(now)
	rp.logger.Println("Measured RTT: ", rtt.String(), "between", addr, "and RP")
	return rtt
}

func (rp *OverlayRP) configureFFRedirecterWithStream(as *availableStream) {
	destinationsAddr := make([]string, 0)
	destinations := make(map[string]*safesockets.SafeUDPWritter)
	for _, requester := range rp.availableStreams[as.name].requesters {
		if _, ok := destinations[requester.node.Addr]; !ok {
			destAddr, _ := net.ResolveUDPAddr("udp", requester.graphPath[1]+":4999")
			conn, _ := net.DialUDP("udp", nil, destAddr)
			destinations[requester.node.Addr] = safesockets.NewSafeUDPWritter(conn)
			destinationsAddr = append(destinationsAddr, requester.node.Addr)
		}
	}
	rp.logger.Println("Changing stream", as.name, "in FFRedirecter with destinations", destinationsAddr)
	rp.redirecter.AddStream(as.name, destinations)
}

func (rp *OverlayRP) getBestServerForStream(as *availableStream) *safesockets.SafeTCPConn {
	bestServer := ""
	bestRtt := time.Duration(math.MaxInt64)
	for serverAddr := range as.hosts {
		rp.servers[serverAddr].rttLock.Lock()
		rttReq := rtt.NewRttPacket(rp.overlay.RPAddr)
		rp.servers[serverAddr].conn.SafeWrite(rttReq)
		rp.servers[serverAddr].rttLock.RLock()
		if rp.servers[serverAddr].rpRtt < bestRtt {
			bestRtt = rp.servers[serverAddr].rpRtt
			bestServer = serverAddr
		}
		rp.servers[serverAddr].rttLock.RUnlock()
	}
	as.pickedHost = bestServer
	rp.logger.Println("Best server for stream", as.name, "is", bestServer, "with RTT", bestRtt)
	return rp.servers[bestServer].conn
}

func (rp *OverlayRP) writeToServer(serverAddr string, p packets.PacketI) {
	rp.mainLock.RLock()
	defer rp.mainLock.RUnlock()
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
	rp.mainLock.RLock()
	defer rp.mainLock.RUnlock()
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
