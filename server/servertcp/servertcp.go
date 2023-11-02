package servertcp
import (
    "log"
    "net"
    "os"
    "time"
    "fmt"
)
func MainTcp(addrr string) {
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
        go handleIncomingRequest(conn)
    }
}
func handleIncomingRequest(conn net.Conn) {
    // store incoming data
    buffer := make([]byte, 1024)
    _, err := conn.Read(buffer)
    if err != nil {
        log.Fatal(err)
    }
    // respond
    time := time.Now().Format("Monday, 02-Jan-06 15:04:05 MST")
    conn.Write([]byte("Hi back!\n"))
    conn.Write([]byte(time))

    // close conn
    conn.Close()
}