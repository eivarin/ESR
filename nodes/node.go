package main
import (
	"fmt"
	"net"
	"os"
	"time"
	"log"
	)
func testcoonection(host string, ports string) {
    
        timeout := time.Second
        conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, ports), timeout)
        if err != nil {
            fmt.Println("Connecting error:", err)
        }
        if conn != nil {
            defer conn.Close()
            fmt.Println("Opened", net.JoinHostPort(host, ports))
        }
    	
}
func ListenTcp(addrr string) {
    fmt.Printf(addrr)
    listen, err := net.Listen("tcp", addrr)
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
    // close listener
    defer listen.Close()
    for {
        conn, err := listen.Accept()
        if err != nil {
            log.Fatal(err)
            os.Exit(1)
        }
		fmt.Println("Connecting:", conn)
    }
}
func main(){
/*
Inicializa:
-Ficar à escuta de conexões
-Tem os vizinhos que indicamos pelos argumentos
-Tentar estabelecer conexão
*/
// --> Isto está confuso, explorar mais
//ListenTcp("10.0.3.2:24")
//testcoonection(os.Args[1],os.Args[2])
}
