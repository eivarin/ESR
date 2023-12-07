package nolonguer

import (
	"bytes"
	"encoding/gob"
)

type NoLonguerAvailablePacket struct {
	StreamName string
}

func NewNoLonguerAvailablePacket(streamName string) *NoLonguerAvailablePacket {
	p := new(NoLonguerAvailablePacket)
	p.StreamName = streamName
	return p
}

func (p *NoLonguerAvailablePacket) Encode() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(p.StreamName)
	return buf.Bytes()
}

func (p *NoLonguerAvailablePacket) Decode(buf []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(&p.StreamName)
}

type NoLonguerInterestedPacket struct {
	StreamName string
	ClientAddress string
}

func NewNoLonguerInterestedPacket(streamName string, clientAddress string) *NoLonguerInterestedPacket {
	p := new(NoLonguerInterestedPacket)
	p.StreamName = streamName
	p.ClientAddress = clientAddress
	return p
}

func (p *NoLonguerInterestedPacket) Encode() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(p.StreamName)
	enc.Encode(p.ClientAddress)
	return buf.Bytes()
}

func (p *NoLonguerInterestedPacket) Decode(buf []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(&p.StreamName)
	dec.Decode(&p.ClientAddress)
}