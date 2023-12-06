package requeststream

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type RequestStreamPacket struct {
	StreamName string
	RequesterAddr string
}

func NewRequestStreamPacket(streamName string, requesterAddr string) *RequestStreamPacket {
	p := RequestStreamPacket{}
	p.StreamName = streamName
	p.RequesterAddr = requesterAddr
	return &p
}

func (p *RequestStreamPacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *RequestStreamPacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}

type RequestStreamResponsePacket struct {
	Status bool
	StreamName string
	Sdp string
}

func NewRequestStreamResponsePacket(status bool, streamName string, sdp string) *RequestStreamResponsePacket {
	p := RequestStreamResponsePacket{}
	p.Status = status
	p.StreamName = streamName
	p.Sdp = sdp
	return &p
}

func (p *RequestStreamResponsePacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *RequestStreamResponsePacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}
