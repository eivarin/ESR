package join

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

type JoinOverlayRPPacket struct {
	OverlayType byte
	LocalAddr string
	NeighborsAddrs []string
	NeighborsRTTs map[string]time.Duration
}

func NewJoinOverlayRPPacket(overlayType byte, localAddr string, neighborsAddrs []string, neighborsRTTs map[string]time.Duration) *JoinOverlayRPPacket {
	p := JoinOverlayRPPacket{}
	p.OverlayType = overlayType
	p.LocalAddr = localAddr
	p.NeighborsAddrs = neighborsAddrs
	p.NeighborsRTTs = neighborsRTTs
	return &p
}

func (p *JoinOverlayRPPacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *JoinOverlayRPPacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}

type JoinOverlayRPResponsePacket struct {
	Success bool
}

func NewJoinOverlayRPResponsePacket(success bool) *JoinOverlayRPResponsePacket {
	p := JoinOverlayRPResponsePacket{}
	p.Success = success
	return &p
}

func (p *JoinOverlayRPResponsePacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *JoinOverlayRPResponsePacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}