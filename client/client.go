package main
import (
	"fmt"
	"net"
	"os"
	)

func main(){
	socket, err := net.ListenPacket("udp","" )
	if err != nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer socket.Close()
	endereco, err := net.ResolveUDPAddr("udp",os.Args[1])
	if err != nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}
	_,err = socket.WriteTo([]byte("mensagem"), endereco)
	if err != nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}
}