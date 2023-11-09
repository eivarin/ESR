package main
import (
	"fmt"
	"net"
    "bufio"
    "strings"
	)
func handleIncomingRequest(c net.Conn) {
    for {

        netData, err := bufio.NewReader(c).ReadString('\n')
        if err != nil {
                return
        }
        if strings.TrimSpace(string(netData)) == "STOP" {
                fmt.Println("Exiting TCP server!")
                return
        }

        fmt.Print("-> ", string(netData))
    }
}
func main(){
        l, err := net.Listen("tcp", "0.0.0.0:8080")
        if err != nil {
                fmt.Println(err)
                return
        }
        defer l.Close()
    for{
        c, err := l.Accept()
        if err != nil {
                fmt.Println(err)
                return
        }
        go handleIncomingRequest(c)
    }
}


