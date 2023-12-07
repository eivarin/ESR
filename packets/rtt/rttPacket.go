package rtt

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type RttPacket struct {
	SenderAddr string
}

func NewRttPacket(senderAddr string) *RttPacket {
	p := RttPacket{}
	p.SenderAddr = senderAddr
	return &p
}

func (p *RttPacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *RttPacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}

type RttResponsePacket struct {
	PeerAddr string
}

func NewRttResponsePacket(peerAddr string) *RttResponsePacket {
	p := RttResponsePacket{}
	p.PeerAddr = peerAddr
	return &p
}

func (p *RttResponsePacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *RttResponsePacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}
