RP_Files= commands.go graphLogic.go main.go
Server_Files= commands.go streams.go main.go
Client_Files= commands.go main.go
Node_Files= commands.go main.go

Path= ./programs

RP_Path= $(Path)/OverlayRP/
Server_Path= $(Path)/OverlayServer/
Client_Path= $(Path)/OverlayClient/
Node_Path=  $(Path)/OverlayNode/

all: RP Server Client Node

RP: 
	cd $(RP_Path) && go build -o RP $(RP_Files)

Server:
	cd $(Server_Path) && go build -o Server $(Server_Files)

Client:
	cd $(Client_Path) && go build -o Client $(Client_Files)

Node:
	cd $(Node_Path) && go build -o Node $(Node_Files)


clean:
	rm -f $(RP_Path)RP $(Server_Path)Server $(Client_Path)Client $(Node_Path)Node


default: all