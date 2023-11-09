package main
import (
	"fmt"
	"net"
	"os"
	)
func udpmessage(addrr string){
	socket, err := net.ListenPacket("udp","")
	if err != nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer socket.Close()
	endereco, err := net.ResolveUDPAddr("udp",addrr)
	if err != nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}
	_,err = socket.WriteTo([]byte("mensagem udp"), endereco)
	if err != nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func tcpmessage(addrr string){
	conn, err := net.Dial("tcp", addrr)
	if err != nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	// send to server
	conn.Write([]byte("Hi back!\n"))
    //fmt.Fprintf(conn, "mensagem tcp")
}
func main(){
	//var udpaddrr string  = os.Args[1]
	var tcpaddrr string = os.Args[2]
	//udpmessage(udpaddrr)
	tcpmessage(tcpaddrr)
}