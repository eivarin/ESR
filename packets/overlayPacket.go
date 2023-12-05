package packets

import (
	"bytes"
	"encoding/gob"
	"main/packets/getstreams"
	"main/packets/hello"
	"main/packets/join"
	"main/packets/publish"
	"main/packets/requeststream"
)

const (
	OverlayPacketTypeHello byte = iota
	OverlayPacketTypeHelloResponse
	OverlayPacketTypeJoin
	OverlayPacketTypeJoinResponse
	OverlayPacketTypePublish
	OverlayPacketTypePublishResponse
	OverlayPacketTypeRequestStream
	OverlayPacketTypeRequestStreamResponse
	OverlayPacketTypeGetStreams
	OverlayPacketTypeGetStreamsResponse
)

type PacketI interface {
	Encode() []byte
	Decode([]byte)
}

type OverlayPacket struct {
	T byte
	D []byte
}

func NewOverlayPacket(d PacketI) *OverlayPacket {
	p := OverlayPacket{}
	switch d.(type) {
	case *hello.JoinOverlayPacket:
		p.T = OverlayPacketTypeHello
	case *hello.JoinOverlayResponsePacket:
		p.T = OverlayPacketTypeHelloResponse
	case *join.JoinOverlayRPPacket:
		p.T = OverlayPacketTypeJoin
	case *join.JoinOverlayRPResponsePacket:
		p.T = OverlayPacketTypeJoinResponse
	case *publish.PublishPacket:
		p.T = OverlayPacketTypePublish
	case *publish.PublishResponsePacket:
		p.T = OverlayPacketTypePublishResponse
	case *requeststream.RequestStreamPacket:
		p.T = OverlayPacketTypeRequestStream
	case *requeststream.RequestStreamResponsePacket:
		p.T = OverlayPacketTypeRequestStreamResponse
	case *getstreams.GetStreamsPacket:
		p.T = OverlayPacketTypeGetStreams
	case *getstreams.GetStreamsResponsePacket:
		p.T = OverlayPacketTypeGetStreamsResponse
	}
	p.D = d.Encode()
	return &p
}

func (p *OverlayPacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (p *OverlayPacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		panic(err)
	}
}

func (p *OverlayPacket) InnerPacket() PacketI {
	switch p.T {
	case OverlayPacketTypeHello:
		hp := hello.JoinOverlayPacket{}
		hp.Decode(p.D)
		return &hp
	case OverlayPacketTypeHelloResponse:
		hp := hello.JoinOverlayResponsePacket{}
		hp.Decode(p.D)
		return &hp
	case OverlayPacketTypeJoin:
		jp := join.JoinOverlayRPPacket{}
		jp.Decode(p.D)
		return &jp
	case OverlayPacketTypeJoinResponse:
		jp := join.JoinOverlayRPResponsePacket{}
		jp.Decode(p.D)
		return &jp
	case OverlayPacketTypePublish:
		pp := publish.PublishPacket{}
		pp.Decode(p.D)
		return &pp
	case OverlayPacketTypePublishResponse:
		pp := publish.PublishResponsePacket{}
		pp.Decode(p.D)
		return &pp
	case OverlayPacketTypeRequestStream:
		rp := requeststream.RequestStreamPacket{}
		rp.Decode(p.D)
		return &rp
	case OverlayPacketTypeRequestStreamResponse:
		rp := requeststream.RequestStreamResponsePacket{}
		rp.Decode(p.D)
		return &rp
	case OverlayPacketTypeGetStreams:
		gsp := getstreams.GetStreamsPacket{}
		gsp.Decode(p.D)
		return &gsp
	case OverlayPacketTypeGetStreamsResponse:
		gsp := getstreams.GetStreamsResponsePacket{}
		gsp.Decode(p.D)
		return &gsp
	}
	return nil
}

func (p *OverlayPacket) Type() byte {
	return p.T
}