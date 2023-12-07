package startstream

import (
	"bytes"
	"encoding/gob"
)

type StartStreamPacket struct {
	StreamName string
	StartTimeSeconds int64
}

func NewStartStreamPacket(streamName string, startTimeSeconds int64) *StartStreamPacket {
	p := new(StartStreamPacket)
	p.StreamName = streamName
	p.StartTimeSeconds = startTimeSeconds
	return p
}

func (p *StartStreamPacket) Encode() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(p.StreamName)
	enc.Encode(p.StartTimeSeconds)
	return buf.Bytes()
}

func (p *StartStreamPacket) Decode(buf []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(&p.StreamName)
	dec.Decode(&p.StartTimeSeconds)
}

type StartStreamResponsePacket struct {
	Success bool
}

func NewStartStreamResponsePacket(success bool) *StartStreamResponsePacket {
	p := new(StartStreamResponsePacket)
	p.Success = success
	return p
}

func (p *StartStreamResponsePacket) Encode() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(p.Success)
	return buf.Bytes()
}

func (p *StartStreamResponsePacket) Decode(buf []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(&p.Success)
}
