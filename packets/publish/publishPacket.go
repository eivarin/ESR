package publish

import (
	"bytes"
	"encoding/gob"
	"fmt"
)


type PublishPacket struct {
	StreamerAddr string
	StreamNames []string
}

func NewPublishPacket(streamerAddr string, streamNames []string) *PublishPacket {
	p := PublishPacket{}
	p.StreamerAddr = streamerAddr
	p.StreamNames = streamNames
	return &p
}

func (p *PublishPacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *PublishPacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}

type PublishResponsePacket struct {
	NumberOfStreamsOk int
}

func NewPublishResponsePacket(numberOfStreamsOk int) *PublishResponsePacket {
	p := PublishResponsePacket{}
	p.NumberOfStreamsOk = numberOfStreamsOk
	return &p
}

func (p *PublishResponsePacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *PublishResponsePacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}
