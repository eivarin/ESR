package main
import (
	"strings"
	"fmt"
	"net"
        "bufio"
        "time"
        "os"
        "log"
        "encoding/json"
        "strconv"
	)
func handleIncomingRequest(c net.Conn) {
    for {

        netData, err := bufio.NewReader(c).ReadString('\n')
        if err != nil {
                return
        }
        //case and handle of msg
        switch {
                case strings.Contains(netData,"Add node:"):
                        ip := string(netData)[9:len(string(netData))-2]
                        addaddrr(ip)
        }
        //
        fmt.Print("-> ", string(netData))
        t := time.Now()
        myTime := t.Format(time.RFC3339) + "\n"
        c.Write([]byte(myTime))
    }
}
func readlist() []map[string]string{
        var list []map[string]string
        file, err := os.Open("addrresslist.txt")
        if err != nil {
        log.Fatal(err)
        }       
        defer file.Close()

        if err := json.NewDecoder(file).Decode(&list); err != nil {
        log.Fatal(err)
        }
        return list
}
func writelist(data []map[string]string){
        jsonStr, err := json.Marshal(data)
        if err != nil {
            fmt.Printf("Error: %s", err.Error())
        } else {
            fmt.Println(string(jsonStr))
        }
        f, err := os.Create("addrresslist.txt")
        if err != nil {
         log.Fatal(err)
        }
        f.WriteString(string(jsonStr))
        f.Sync()
}
func addaddrr(value string){
        var data []map[string]string = readlist()
        newMap := make(map[string]string)
        newMap["n" + strconv.Itoa(len(data)+1)] = value
        data = append(data, newMap) 
        writelist(data)
}
func main(){
        //eventualmente ter vários ficheiros para cada nó,
        // cada um terá a sua lista de vizinhos otherwise não faz sentido
        var data []map[string]string = readlist()
       
        /*
        Contando que cada nó tem uma lista individual:
        -Ve se a lista têm elementos
        -Se tiver funcionamento normal
                -Se não tiver, manda um multicast para 
                 todas as interfaces, recebendo respostas.
        */
       
        writelist(data)
        //------x------
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



