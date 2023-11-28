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
        fmt.Println(err)
    }
	return buf.Bytes()
}

func decodeMediaPacket(packet []byte) *MediaPacket {
    dec := gob.NewDecoder(bytes.NewBuffer(packet))
	parsed_packet := new(MediaPacket)
    if err := dec.Decode(parsed_packet); err != nil {
        fmt.Println(err)
    }
	return parsed_packet
}

// func test_mediaPacket() {
// 	s1 := MediaPacket{"Hello, World!", 0, 42, []byte("This is a test")}
// 	encoded_packet := encodeMediaPacket(s1)
// 	ds1 := decodeMediaPacket(encoded_packet)
// 	fmt.Printf("%s" , ds1.Payload)
// }
