package overlay

import (
	"fmt"
	"io"
	"log"
	"main/packets"
	"main/packets/hello"
	"main/safesockets"
	"net"
	"sync"
	"time"
)

const (
	NodeOverlay byte = iota
	NodeClient
	NodeServer
	NodeRP
)

type NodeI interface {
	JoinOverlay(rpAddr *net.TCPAddr, localAddr string, neighborsAddrs []string, neighborsRTTs map[string]time.Duration)
	LeaveOverlay()
	HandlePacketFromNeighbor(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn)
	HandlePacketFromRP(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn)
}
type Neighbor struct {
	logger   *log.Logger
	Conn     *safesockets.SafeTCPConn
	NodeType byte
	RTT      time.Duration
}

type Overlay struct {
	logger        *log.Logger
	Neighbors     map[string]*Neighbor
	NeighborsLock sync.RWMutex
	RttLock       *sync.Mutex
	RPAddr        string
	basePort      int
	overlayType   byte
	LocalAddr     string
}

func NewOverlay(logger *log.Logger, basePort int, overlayType byte, localAddr string, neighbors []string, node NodeI) *Overlay {
	overlay := new(Overlay)
	overlay.logger = logger
	overlay.Neighbors = make(map[string]*Neighbor)
	overlay.basePort = basePort
	overlay.overlayType = overlayType
	overlay.RttLock = &sync.Mutex{}
	overlay.LocalAddr = localAddr
	for _, neighbor := range neighbors {
		overlay.hiNeighbor(neighbor, node.HandlePacketFromNeighbor)
	}
	addr, _ := net.ResolveTCPAddr("tcp", overlay.RPAddr+":"+fmt.Sprint(overlay.basePort))
	rttList := make(map[string]time.Duration, len(overlay.Neighbors))
	for i, neigh := range overlay.Neighbors {
		rttList[i] = neigh.RTT
	}
	node.JoinOverlay(addr, localAddr, neighbors, rttList)
	if overlayType == NodeOverlay || overlayType == NodeRP {
		go overlay.listenForNewNeighbors(node.HandlePacketFromNeighbor)
	}
	return overlay
}

func NewOverlayRP(logger *log.Logger, basePort int, rpAddr string, node NodeI) *Overlay {
	overlay := new(Overlay)
	overlay.logger = logger
	overlay.Neighbors = make(map[string]*Neighbor)
	overlay.basePort = basePort
	overlay.overlayType = NodeRP
	overlay.RPAddr = rpAddr
	overlay.LocalAddr = rpAddr
	overlay.RttLock = &sync.Mutex{}
	go overlay.listenForNewNeighbors(node.HandlePacketFromNeighbor)
	return overlay
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func (o *Overlay) measureRTTFromClient(listenningConn net.Conn, writtingConn net.Conn, addr string, localAddr string) time.Duration {
	defer writtingConn.Close()
	defer listenningConn.Close()

	buf := make([]byte, 10)
	now := time.Now()

	writtingConn.Write([]byte("rtt"))
	listenningConn.Read(buf)
	rtt := time.Since(now)
	o.logger.Println("Measured RTT: ", rtt.String(), "between", addr, "and", localAddr)
	return rtt
}

func (o *Overlay) measureRTTFromServer(listenningConn net.Conn, writtingConn net.Conn) {
	defer writtingConn.Close()
	defer listenningConn.Close()

	buf := make([]byte, 10)
	listenningConn.Read(buf)
	writtingConn.Write([]byte("rtt"))
	o.RttLock.Unlock()
}

func (o *Overlay) hiNeighbor(addr string, handleNewPacket func(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn)) {
	// connect to neighbor
	ad, _ := net.ResolveTCPAddr("tcp", addr+":"+fmt.Sprint(o.basePort+1))
	conn, err := net.DialTCP("tcp", nil, ad)
	if err != nil {
		o.logger.Println("Error connecting to neighbor", addr, err)
		return
	}

	// send hello packet
	op := packets.NewOverlayPacket(hello.NewJoinOverlayPacket(o.overlayType, o.LocalAddr))
	payload := op.Encode()
	conn.Write(payload)

	// read hello response packet
	buf := make([]byte, 1500)
	nBytes, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			o.logger.Println("Neighbor", addr, "closed connection")
			return
		}
		o.logger.Println("Error reading from neighbor", addr, err)
		return
	}
	o.RttLock.Lock()
	lad, _ := net.ResolveUDPAddr("udp", o.LocalAddr+":"+fmt.Sprint(o.basePort-4))
	listenningConn, _ := net.ListenUDP("udp", lad)
	wad, _ := net.ResolveUDPAddr("udp", addr+":"+fmt.Sprint(o.basePort-4))
	writtingConn, _ := net.DialUDP("udp", nil, wad)
	rtt := o.measureRTTFromClient(listenningConn, writtingConn, addr, o.LocalAddr)
	o.RttLock.Unlock()
	// decode hello response packet
	opResp := packets.OverlayPacket{}
	opResp.Decode(buf[:nBytes])
	hp := opResp.InnerPacket().(*hello.HelloOverlayResponsePacket)

	// add neighbor to neighbors
	neigh := NewNeighbor(conn, hp.OverlayType, o.logger)
	neigh.RTT = rtt
	o.NeighborsLock.Lock()
	o.Neighbors[neigh.getNeighborIp()] = neigh
	o.NeighborsLock.Unlock()
	go neigh.handle(handleNewPacket)
	o.RPAddr = hp.RPAddr
}

func NewNeighbor(conn *net.TCPConn, t byte, logger *log.Logger) *Neighbor {
	neighbor := new(Neighbor)
	neighbor.Conn = safesockets.NewSafeTCPConn(conn)
	neighbor.NodeType = t
	neighbor.logger = logger
	return neighbor
}

func (neigh *Neighbor) handle(handleNewPacket func(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn)) {
	for {
		p, t, _, err := neigh.Conn.SafeRead()
		if err != nil {
			if err == io.EOF {
				neigh.logger.Println("Neighbor", neigh.getNeighborIp(), "closed connection")
				return
			}
			neigh.logger.Println("Error reading from neighbor", neigh.getNeighborIp(), err)
			continue
		}
		go handleNewPacket(p, t, neigh.Conn)
	}
}

func (neigh *Neighbor) getNeighborIp() string {
	return neigh.Conn.GetRemoteAddr()
}

func (o *Overlay) listenForNewNeighbors(handleNewPacket func(p packets.PacketI, t byte, conn *safesockets.SafeTCPConn)) {
	addr, _ := net.ResolveTCPAddr("tcp", ":"+fmt.Sprint(o.basePort+1))
	listener, _ := net.ListenTCP("tcp", addr)
	for {
		// accept new neighbor
		conn, err := listener.AcceptTCP()
		if err != nil {
			o.logger.Println("Error accepting new neighbor", err)
			continue
		}
		// new neighbor
		go func(conn *net.TCPConn) {
			// read hello packet
			buf := make([]byte, 1500)
			nBytes, err := conn.Read(buf)
			if err != nil {
				o.logger.Println("Error reading from neighbor", conn.RemoteAddr().(*net.TCPAddr).IP.String(), err)
				return
			}
			// decode hello packet
			op := packets.OverlayPacket{}
			op.Decode(buf[:nBytes])
			hp := op.InnerPacket().(*hello.HelloOverlayPacket)
			o.RttLock.Lock()
			lad, _ := net.ResolveUDPAddr("udp", o.LocalAddr+":"+fmt.Sprint(o.basePort-4))
			listenningConn, _ := net.ListenUDP("udp", lad)
			wad, _ := net.ResolveUDPAddr("udp", hp.LocalAddr+":"+fmt.Sprint(o.basePort-4))
			writtingConn, _ := net.DialUDP("udp", nil, wad)
			go o.measureRTTFromServer(listenningConn, writtingConn)

			// send hello response packet
			opResp := packets.NewOverlayPacket(hello.NewJoinOverlayResponsePacket(hp.OverlayType, o.RPAddr))
			payload := opResp.Encode()
			conn.Write(payload)

			// add neighbor to neighbors
			neigh := NewNeighbor(conn, hp.OverlayType, o.logger)
			o.NeighborsLock.Lock()
			o.Neighbors[neigh.getNeighborIp()] = neigh
			o.NeighborsLock.Unlock()
			go neigh.handle(handleNewPacket)
		}(conn)
	}
}
