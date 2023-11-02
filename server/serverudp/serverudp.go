package serverudp
import (
"fmt"
"net"
"os"
)

func MainUdp(addrr string){
	fmt.Println("Running")

	socket, err := net.ListenPacket("udp", addrr)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer socket.Close()
	
	fmt.Printf("Open UDP Socket on: %s \n", addrr)

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
