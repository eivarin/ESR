package main
import (
	"fmt"
	"net"
        "bufio"
        "time"
        "os"
        "log"
        "encoding/json"
	)
func handleIncomingRequest(c net.Conn) {
    for {

        netData, err := bufio.NewReader(c).ReadString('\n')
        if err != nil {
                return
        }

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
func main(){
        //Nota de que no estado atual não conseguimos adicionar mais nós ao mapa
        var data []map[string]string = readlist()
        data["n3"]= "10.0.2.2:8080"
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



