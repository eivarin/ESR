package main
import (
"fmt"
"net"
"os"
)

func main(){
	fmt.Println("Running")

	socket, err := net.ListenPacket("udp", os.Args[1])

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer socket.Close()
	
	fmt.Printf("Open Socket on: %s", os.Args[1])

	buffer := make([]byte, 1024)

	for {
	_, remetente, err := socket.ReadFrom(buffer)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("Mensagem do %s: %s\n\n", remetente.String(), string(buffer))
	}
}
