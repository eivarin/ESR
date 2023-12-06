package getstreams

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type GetStreamsPacket struct {
	Success bool
}

func NewGetStreamsPacket(success bool) *GetStreamsPacket {
	p := GetStreamsPacket{}
	p.Success = success
	return &p
}

func (p *GetStreamsPacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *GetStreamsPacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}

type GetStreamsResponsePacket struct {
	StreamNames []string
}

func NewGetStreamsResponsePacket(streamNames []string) *GetStreamsResponsePacket {
	p := GetStreamsResponsePacket{}
	p.StreamNames = streamNames
	return &p
}

func (p *GetStreamsResponsePacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *GetStreamsResponsePacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}