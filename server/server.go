package main
import (
	"os"
	stcp "server/servertcp"
	sudp "server/serverudp"
	)
func main(){
sudp.MainUdp(os.Args[1])
stcp.MainTcp(os.Args[2])
}