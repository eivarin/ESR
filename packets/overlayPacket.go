package packets

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"main/packets/getstreams"
	"main/packets/hello"
	"main/packets/join"
	"main/packets/publish"
	"main/packets/ready"
	"main/packets/redirects"
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
	OverlayPacketTypeRedirects
	OverlayPacketTypeRedirectsResponse
	OverlayPacketTypeReady
	OverlayPacketTypeReadyResponse
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
	case *hello.HelloOverlayPacket:
		p.T = OverlayPacketTypeHello
	case *hello.HelloOverlayResponsePacket:
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
	case *redirects.RedirectsPacket:
		p.T = OverlayPacketTypeRedirects
	case *redirects.RedirectsResponsePacket:
		p.T = OverlayPacketTypeRedirectsResponse
	case *ready.ReadyPacket:
		p.T = OverlayPacketTypeReady
	case *ready.ReadyResponsePacket:
		p.T = OverlayPacketTypeReadyResponse
	}
	p.D = d.Encode()
	return &p
}

func (p *OverlayPacket) Encode() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		fmt.Println("Error encoding packet")
		panic(err)
	}
	return buf.Bytes()
}

func (p *OverlayPacket) Decode(packet []byte) {
	dec := gob.NewDecoder(bytes.NewBuffer(packet))
	if err := dec.Decode(p); err != nil {
		fmt.Println("Error decoding packet")
		panic(err)
	}
}

func (p *OverlayPacket) InnerPacket() PacketI {
	var packet PacketI
	switch p.T {
	case OverlayPacketTypeHello:
		packet = new(hello.HelloOverlayPacket)
	case OverlayPacketTypeHelloResponse:
		packet = new(hello.HelloOverlayResponsePacket)
	case OverlayPacketTypeJoin:
		packet = new(join.JoinOverlayRPPacket)
	case OverlayPacketTypeJoinResponse:
		packet = new(join.JoinOverlayRPResponsePacket)
	case OverlayPacketTypePublish:
		packet = new(publish.PublishPacket)
	case OverlayPacketTypePublishResponse:
		packet = new(publish.PublishResponsePacket)
	case OverlayPacketTypeRequestStream:
		packet = new(requeststream.RequestStreamPacket)
	case OverlayPacketTypeRequestStreamResponse:
		packet = new(requeststream.RequestStreamResponsePacket)
	case OverlayPacketTypeGetStreams:
		packet = new(getstreams.GetStreamsPacket)
	case OverlayPacketTypeGetStreamsResponse:
		packet = new(getstreams.GetStreamsResponsePacket)
	case OverlayPacketTypeRedirects:
		packet = new(redirects.RedirectsPacket)
	case OverlayPacketTypeRedirectsResponse:
		packet = new(redirects.RedirectsResponsePacket)
	case OverlayPacketTypeReady:
		packet = new(ready.ReadyPacket)
	case OverlayPacketTypeReadyResponse:
		packet = new(ready.ReadyResponsePacket)
	}
	packet.Decode(p.D)
	return packet
}

func (p *OverlayPacket) Type() byte {
	return p.T
}
