package main
import (
	"fmt"
	"net"
	"os"
	"bufio"
	"strings"
	)

func tcpmessage(addrr string){
	conn, err := net.Dial("tcp", addrr)
	if err != nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	// send to server
	conn.Write([]byte("Hi back!\n"))
}
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
        
        arguments := os.Args
        if len(arguments) == 1 {
                fmt.Println("Please provide host:port.")
                return
        }

        CONNECT := arguments[1]
        c, err := net.Dial("tcp", CONNECT)
        if err != nil {
                fmt.Println(err)
                return
        }

        for {
                reader := bufio.NewReader(os.Stdin)
                fmt.Print(">> ")
                text, _ := reader.ReadString('\n')
                fmt.Fprintf(c, text+"\n")

                message, _ := bufio.NewReader(c).ReadString('\n')
                fmt.Print("->: " + message)
        }
}