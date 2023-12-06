package redirects

import (
	"bytes"
	"encoding/gob"
)

type RedirectsPacket struct {
	StreamName string
	Redirects []string
}

func NewRedirectsPacket(streamName string, redirects []string) *RedirectsPacket {
	p := new(RedirectsPacket)
	p.StreamName = streamName
	p.Redirects = redirects
	return p
}

func (p *RedirectsPacket) Encode() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(p.StreamName)
	enc.Encode(p.Redirects)
	return buf.Bytes()
}

func (p *RedirectsPacket) Decode(buf []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(&p.StreamName)
	dec.Decode(&p.Redirects)
}

type RedirectsResponsePacket struct {
	Success bool
}

func NewRedirectsResponsePacket(success bool) *RedirectsResponsePacket {
	p := new(RedirectsResponsePacket)
	p.Success = success
	return p
}

func (p *RedirectsResponsePacket) Encode() []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(p.Success)
	return buf.Bytes()
}

func (p *RedirectsResponsePacket) Decode(buf []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(&p.Success)
}