package ffmpeg

import (
	"log"
	"net"
	"sync"
	"main/safesockets"
)



type stream struct {
	Name string
	destinyConns map[string]*safesockets.SafeUDPWritter
	writtingLock *sync.Mutex
}

func newStream(name string, destinyConns map[string]*safesockets.SafeUDPWritter) *stream {
	s := new(stream)
	s.Name = name
	s.destinyConns = destinyConns
	s.writtingLock = new(sync.Mutex)
	return s
}

type FFRedirecter struct {
	logger *log.Logger
	recConn *safesockets.SafeUDPReader
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
		logger.Println("Error listening UDP", err)
		log.Panic(err)
	}
	ffr.recConn = safesockets.NewSafeUDPReader(recConn)
	ffr.streams = make(map[string]*stream)
	ffr.streamsLock = new(sync.RWMutex)
	go func() {
		for {
			buf, n, err := ffr.recConn.SafeRead()
			if err != nil {
				logger.Println(err)
			}
			mPacket := decodeMediaPacket(buf[:n])
			ffr.streamsLock.RLock()
			if s, ok := ffr.streams[mPacket.StreamName]; ok {
				s.writtingLock.Lock()
				// ffr.logger.Println(mPacket.StreamName, s.destinyConns)
				for _, conn := range s.destinyConns {
					conn.SafeWrite(buf)
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

func (ffr *FFRedirecter) AddStream(name string, destinyConns map[string]*safesockets.SafeUDPWritter) {
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