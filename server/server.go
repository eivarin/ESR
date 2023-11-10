package main
import (
	"os"
	stcp "main/server/servertcp"
	sudp "main/server/serverudp"
	)
func main(){
sudp.MainUdp(os.Args[1])
stcp.MainTcp(os.Args[2])
}