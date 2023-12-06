package safesockets

import (
	"main/packets"
	"net"
	"sync"
)

type SafeTCPConn struct {
	conn *net.TCPConn
	writeLock *sync.Mutex
	readLock *sync.Mutex
}

func NewSafeTCPConn(conn *net.TCPConn) *SafeTCPConn {
	safeConn := new(SafeTCPConn)
	safeConn.conn = conn
	safeConn.writeLock = &sync.Mutex{}
	safeConn.readLock = &sync.Mutex{}
	return safeConn
}

func (safeConn *SafeTCPConn) SafeRead() (p packets.PacketI, t byte, n int, err error) {
	buf := make([]byte, 1500)
	safeConn.readLock.Lock()
	n, err = safeConn.conn.Read(buf)
	if err != nil {
		return nil, 0, 0, err
	}
	safeConn.readLock.Unlock()
	op := packets.OverlayPacket{}
	op.Decode(buf[:n])
	t = op.Type()
	p = op.InnerPacket()
	return
}

func (safeConn *SafeTCPConn) SafeWrite(p packets.PacketI) (n int, err error) {
	OverlayPacket := packets.NewOverlayPacket(p)
	safeConn.writeLock.Lock()
	n, err = safeConn.conn.Write(OverlayPacket.Encode())
	safeConn.writeLock.Unlock()
	return
}

func (conn *SafeTCPConn) GetRemoteAddr() string {
	return conn.conn.RemoteAddr().(*net.TCPAddr).IP.String()
}


type SafeUDPWritter struct {
	conn *net.UDPConn
	writeLock *sync.Mutex
}

func NewSafeUDPWritter(conn *net.UDPConn) *SafeUDPWritter {
	safeWritter := new(SafeUDPWritter)
	safeWritter.conn = conn
	safeWritter.writeLock = &sync.Mutex{}
	return safeWritter
}

func (safeWritter *SafeUDPWritter) SafeWrite(buf []byte) (n int, err error) {
	safeWritter.writeLock.Lock()
	n, err = safeWritter.conn.Write(buf)
	safeWritter.writeLock.Unlock()
	return
}

type SafeUDPReader struct {
	conn *net.UDPConn
	readLock *sync.Mutex
}

func NewSafeUDPReader(conn *net.UDPConn) *SafeUDPReader {
	safeReader := new(SafeUDPReader)
	safeReader.conn = conn
	safeReader.readLock = &sync.Mutex{}
	return safeReader
}

func (safeReader *SafeUDPReader) SafeRead() (buf []byte, n int, err error) {
	buf = make([]byte, 2000)
	n, err = safeReader.conn.Read(buf)
	return
}
