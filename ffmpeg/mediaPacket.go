package ffmpeg

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type MediaPacket struct {
	StreamName string
	Type byte
	PayloadSize int
	Payload []byte
}

func encodeMediaPacket(packet MediaPacket) []byte {
	buf := bytes.Buffer{}
    enc := gob.NewEncoder(&buf)
    if err := enc.Encode(packet); err != nil {
		fmt.Println("Error encoding media packet")
        fmt.Println(err)
    }
	return buf.Bytes()
}

func decodeMediaPacket(packet []byte) *MediaPacket {
    dec := gob.NewDecoder(bytes.NewBuffer(packet))
	parsed_packet := new(MediaPacket)
    if err := dec.Decode(parsed_packet); err != nil {
		fmt.Println("Error decoding media packet")
        fmt.Println(err)
    }
	return parsed_packet
}