package hello

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type HelloOverlayPacket struct {
	OverlayType byte
	LocalAddr   string
}

func NewJoinOverlayPacket(overlayType byte, localAddr string) *HelloOverlayPacket {
	p := HelloOverlayPacket{}
	p.OverlayType = overlayType
	p.LocalAddr = localAddr
	return &p
}

func (p *HelloOverlayPacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *HelloOverlayPacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}

type HelloOverlayResponsePacket struct {
	OverlayType byte
	RPAddr      string
}

func NewJoinOverlayResponsePacket(overlayType byte, rpaddr string) *HelloOverlayResponsePacket {
	p := HelloOverlayResponsePacket{}
	p.OverlayType = overlayType
	p.RPAddr = rpaddr
	return &p
}

func (p *HelloOverlayResponsePacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *HelloOverlayResponsePacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}
