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
                        //ip := string(netData)[9:len(string(netData))-2]
                        //addaddrr(ip)
        }
        //
        fmt.Print("-> ", string(netData))
        t := time.Now()
        myTime := t.Format(time.RFC3339) + "\n"
        c.Write([]byte(myTime))
    }
 }
func readlist() []map[string]string {
	var list []map[string]string
	// Specify the file path
	filePath := "/home/core/TP/ESR/cmd/AddrDB/" + os.Args[1] + "addrresslist.txt"
	// Attempt to open the file
	file, err := os.Open(filePath)
	if err != nil {
		// If the file doesn't exist, create it
		if os.IsNotExist(err) {
			fmt.Printf("File %s doesn't exist. Creating...\n", filePath)
			file, err = os.Create(filePath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
		} else {
			// If there's another error, log it and exit
			log.Fatal(err)
		}
	} else {
		// If the file was opened successfully, close it
		defer file.Close()
		// Decode the JSON data into the list
		if err := json.NewDecoder(file).Decode(&list); err != nil {
			log.Fatal(err)
		}
	}
	return list
 }
func writelist(data []map[string]string){
        jsonStr, err := json.Marshal(data)
        if err != nil {
            fmt.Printf("Error: %s", err.Error())
        } else {
            fmt.Printf("Write of know nodes from %s sucessfull!\n", os.Args[1])
        }
        f, err := os.Create("/home/core/TP/ESR/cmd/AddrDB/" + os.Args[1] + "addrresslist.txt")
        if err != nil {
         log.Fatal(err)
        }
        f.WriteString(string(jsonStr))
        f.Sync()
 }
func addNode(value string, node string){
        var data []map[string]string = readlist()
        newMap := make(map[string]string)
        newMap["n" + strconv.Itoa(len(data)+1)] = value
        data = append(data, newMap) 
        writelist(data)
}

func listeningtoNeighbour(){
        fmt.Printf("Listening Neighbours\n")
        conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: 12345,
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
       
	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Received message from neighbour %s: %s\n", addr.String(), string(buffer[:n]))
                //case to see if node exists in "db" 
                switch {
                case strings.Contains(string(buffer[:n]),"Hello, Neighbour! I'm node "):
                        if !checkNode(addr.String()){
                                node := string(string(buffer[:n]))[27:len(string(buffer[:n]))]
                                addNode(addr.String(), node)
                        }
        }
   }
        //neighbour must say who he is when he responds to a hello neighbour
}
func checkNode(x string) bool {
	data := readlist()

	// Traverse through the slice of maps
	for _, value := range data {
		// Iterate over each map in the slice
		for _, v := range value {
			// Check if present value is equal to userValue
			if v == x {
				// If the same, return true
				return true
			}
		}
	}

	// If the value is not found, return false
	return false
        }
func getNeighbours(){
        fmt.Printf("Go get Neighbours\n")
        message := []byte("Hello, Neighbours! I'm node " + os.Args[1])
                // Get a list of network interfaces
                interfaces, err := net.Interfaces()
                if err != nil {
                        fmt.Println("Error:", err)
                        return
                }
                // Iterate over each network interface
                for _, iface := range interfaces {
                        // Check if the interface is up and not a loopback interface
                        if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
                                addrs, err := iface.Addrs()
                                if err != nil {
                                        fmt.Println("Error:", err)
                                        continue
                                }
                                // Iterate over each address associated with the interface
                                for _, addr := range addrs {
                                        switch v := addr.(type) {
                                        case *net.IPNet:
                                                // Check if it's an IPv4 address
                                                if v.IP.To4() != nil {
                                                        broadcastAddr := getBroadcastAddress(v)
                                                        go sendBroadcast(broadcastAddr, message)
                                                }
                                        }
                                }
                        }
                }
                // Sleep to allow time for messages to be sent
                time.Sleep(5 * time.Second)
        }
        
func getBroadcastAddress(ipnet *net.IPNet) string {
                ip := ipnet.IP.Mask(ipnet.Mask)
                ip[3] = 255 // Set the host part to 255 for the broadcast address
                return ip.String()
        }
func sendBroadcast(broadcastAddr string, message []byte) {
                conn, err := net.Dial("udp", broadcastAddr+":12345")
                if err != nil {
                        fmt.Println("Error:", err)
                        return
                }
                defer conn.Close()
        
                _, err = conn.Write(message)
                if err != nil {
                        fmt.Println("Error:", err)
                        return
                }
        }

func main(){
        //eventualmente ter vários ficheiros para cada nó,
        
        // cada um terá a sua lista de vizinhos otherwise não faz sentido
        
        //recebe o seu nome pela cli
        fmt.Printf("Initializing node: %s\n", os.Args[1])
        go listeningtoNeighbour();//inicializa e fica logo à escuta de vizinhos
        var data []map[string]string = readlist() // le a lista dos nós que conhece
        if len(data) == 0{
                go getNeighbours();//se a lista estiver vazia pede os ips dos vizinhos
        }
        /*
        Contando que cada nó tem uma lista individual:
        -Ve se a lista têm elementos
        -Se tiver: funcionamento normal
                -Se não tiver, manda um multicast para 
                 todas as interfaces, recebendo respostas e ficando apenas com os vizinhos.
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







