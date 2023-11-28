package shared

import (
	"errors"
	"net"
	"syscall"
	"fmt"
)

func testPort(port int) bool {
	RemoteAddr, _ := net.ResolveUDPAddr("udp", ":"+fmt.Sprint(port))
	c, err := net.ListenUDP("udp", RemoteAddr)
	if err == nil {
		c.Close()
	}
	return errors.Is(err, syscall.EADDRINUSE) || errors.Is(err, syscall.EADDRNOTAVAIL)
}

func GetPorts(basePort int) []int {
	ports := make([]int, 4)
	newPort := basePort
	if newPort%2 == 1 {
		newPort++
	}
	for {
		if !testPort(newPort) || !testPort(newPort+1) {
			ports[0] = newPort
			ports[1] = newPort + 1
			newPort += 2
			break
		}
		newPort += 2
	}
	for {
		if !testPort(newPort) || !testPort(newPort+1) {
			ports[2] = newPort
			ports[3] = newPort + 1
			newPort += 2
			break
		}
		newPort += 2
	}
	return ports
}