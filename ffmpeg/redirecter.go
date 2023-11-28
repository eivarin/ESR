package ffmpeg

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

func Open_Streamer_Sockets(logger *log.Logger, basePort int, remotePort int, destinyIP string, videoName string) []*net.UDPConn {
	ffmpegIP := "127.0.0.1"
	ffmpegCons := make([]*net.UDPConn, 4)
	for i := 0; i < 4; i++ {
		RemoteAddr, _ := net.ResolveUDPAddr("udp", ffmpegIP+":"+fmt.Sprint(basePort+i))
		conn, err := net.ListenUDP("udp", RemoteAddr)
		if err != nil {
			logger.Fatal(err)
		}
		defer conn.Close()
		ffmpegCons[i] = conn
	}

	RemoteAddr, _ := net.ResolveUDPAddr("udp", destinyIP+":"+fmt.Sprint(remotePort))
	SendConn, _ := net.DialUDP("udp", nil, RemoteAddr)
	mu := sync.Mutex{}
	for i := 0; i < 4; i++ {
		go func(tp int) {
			for {
				buf := make([]byte, 2000)
				n, err := ffmpegCons[tp].Read(buf)
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						break
					} else {
						logger.Fatal(err)
					}
				}
				go func() {
					wrappedPacket := encodeMediaPacket(MediaPacket{videoName, byte(tp), n, buf[:n]})
					mu.Lock()
					SendConn.Write(wrappedPacket)
					mu.Unlock()
				}()
			}
		}(i)
	}
	return ffmpegCons
}

//unnecessary
// func open_Player_Sockets(logger *log.Logger, basePort int, remotePort int) *net.UDPConn {
// 	ffplayCons := make([]*net.UDPConn, 4)
// 	ffplayAddr := "127.0.0.1"
// 	for i := 0; i < 4; i++ {
// 		RemoteAddr, _ := net.ResolveUDPAddr("udp", ffplayAddr+":"+fmt.Sprint(basePort+i))
// 		conn, err := net.DialUDP("udp", nil, RemoteAddr)
// 		if err != nil {
// 			logger.Fatal(err)
// 		}
// 		defer conn.Close()
// 		ffplayCons[i] = conn
// 	}
		
// 	incommingAddr, err := net.ResolveUDPAddr("udp","0.0.0.0:"+fmt.Sprint(remotePort))
// 	if err != nil {
// 		logger.Fatal(err)
// 	}
// 	connIncomming, err := net.ListenUDP("udp", incommingAddr)
// 	if err != nil {
// 		logger.Fatal(err)
// 	}
	
// 	go func() {
// 		for {
// 			buf := make([]byte, 2000)
// 			n, err := connIncomming.Read(buf)
// 			if err != nil {
// 				if errors.Is(err, net.ErrClosed) {
// 					break
// 				} else {
// 					logger.Fatal(err)
// 				}
// 			}
// 			go func() {
// 				packet := decodeMediaPacket(buf[:n])
// 				ffplayCons[packet.Type].Write(packet.Payload)
// 			}()
// 		}
// 	}()
// 	return connIncomming
// }