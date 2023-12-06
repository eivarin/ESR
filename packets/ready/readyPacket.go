package ready

import (
	"bytes"
	"encoding/gob"
)

type ReadyPacket struct {
	StreamName string
}

func NewReadyPacket(streamName string) *ReadyPacket {
	p := new(ReadyPacket)
	p.StreamName = streamName
	return p
}

func (p *ReadyPacket) Encode() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(p.StreamName)
	return buf.Bytes()
}

func (p *ReadyPacket) Decode(buf []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(&p.StreamName)
}

type ReadyResponsePacket struct {
	Success bool
}

func NewReadyResponsePacket(success bool) *ReadyResponsePacket {
	p := new(ReadyResponsePacket)
	p.Success = success
	return p
}

func (p *ReadyResponsePacket) Encode() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(p.Success)
	return buf.Bytes()
}

func (p *ReadyResponsePacket) Decode(buf []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(&p.Success)
}