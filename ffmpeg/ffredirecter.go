package ffmpeg

import (
	"log"
	"net"
	"sync"
)



type stream struct {
	Name string
	destinyConns map[string]*net.UDPConn
	writtingLock *sync.Mutex
}

func newStream(name string, destinyConns map[string]*net.UDPConn) *stream {
	s := new(stream)
	s.Name = name
	s.destinyConns = destinyConns
	s.writtingLock = new(sync.Mutex)
	return s
}

type FFRedirecter struct {
	logger *log.Logger
	recConn *net.UDPConn
	streams map[string]*stream
	streamsLock *sync.RWMutex
	debug bool
}

func NewFFRedirecter(logger *log.Logger, listeningUdpAddr *net.UDPAddr, debug bool) *FFRedirecter {
	ffr := new(FFRedirecter)
	ffr.debug = debug
	ffr.logger = logger
	recConn, err := net.ListenUDP("udp", listeningUdpAddr)
	if err != nil {
		log.Panic(err)
	}
	ffr.recConn = recConn
	ffr.streams = make(map[string]*stream)
	ffr.streamsLock = new(sync.RWMutex)
	go func() {
		for {
			buf := make([]byte, 1500)
			n, err := ffr.recConn.Read(buf)
			if err != nil {
				log.Panic(err)
			}
			mPacket := decodeMediaPacket(buf[:n])
			ffr.streamsLock.RLock()
			if s, ok := ffr.streams[mPacket.StreamName]; ok {
				s.writtingLock.Lock()
				for _, conn := range s.destinyConns {
					conn.Write(buf)
				}
				s.writtingLock.Unlock()
			} else if ffr.debug {
				ffr.logger.Printf("Stream %s not found\n", mPacket.StreamName)
			}
			ffr.streamsLock.RUnlock()
		}
	}()
	return ffr
}

func (ffr *FFRedirecter) AddStream(name string, destinyConns map[string]*net.UDPConn) {
	ffr.streamsLock.Lock()
	defer ffr.streamsLock.Unlock()
	s, ok := ffr.streams[name]
	if !ok {
		s = newStream(name, destinyConns)
		ffr.streams[name] = s
	} else {
		s.destinyConns = destinyConns
	}
}

func (ffr *FFRedirecter) RemoveStream(name string) {
	ffr.streamsLock.Lock()
	delete(ffr.streams, name)
	ffr.streamsLock.Unlock()
}