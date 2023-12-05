package hello

import (
	"bytes"
	"encoding/gob"
)

type JoinOverlayPacket struct {
	OverlayType byte
	LocalAddr string
}

func NewJoinOverlayPacket(overlayType byte, localAddr string) *JoinOverlayPacket {
	p := JoinOverlayPacket{}
	p.OverlayType = overlayType
	p.LocalAddr = localAddr
	return &p
}

func (p *JoinOverlayPacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (p *JoinOverlayPacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		panic(err)
	}
}

type JoinOverlayResponsePacket struct {
	OverlayType byte
	RPAddr string
}

func NewJoinOverlayResponsePacket(overlayType byte, rpaddr string) *JoinOverlayResponsePacket {
	p := JoinOverlayResponsePacket{}
	p.OverlayType = overlayType
	p.RPAddr = rpaddr
	return &p
}

func (p *JoinOverlayResponsePacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (p *JoinOverlayResponsePacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		panic(err)
	}
}
