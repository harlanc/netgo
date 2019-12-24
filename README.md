# NetGo


 [![Go Report Card][3]][4] 


[3]: https://goreportcard.com/badge/github.com/netgo-framework/netgo
[4]: https://goreportcard.com/report/github.com/netgo-framework/netgo


Netgo is a Flexible,Powerful,Friendly network synchronization engine,it can be used for VR Application.And new features are under development..



## FeatureList

- [x] Support Room concept
- [x] Support multiple synchronization ways,including RPC and view sync
- [x] Support custom event
- [ ] Support Lobby concept
- [ ] Support Load Balance
- [ ] ...

## Start Server	

### Clone 

Issue the following command to clone the server codes to local:

    git clone https://github.com/netgo-framework/netgo.git
     
### Get dependencies

Go to the netgo root folder and issue the following commad:    
      
    go get -d ./...
    
### Change IP
Open [main.go](https://github.com/netgo-framework/netgo/blob/master/main.go) and  update the ip and port:

    tcpAddr, err := net.ResolveTCPAddr("tcp", "192.168.0.104:8686")
    

### Start Netgo

    go run main.go
    
## How to Run Client Demo

    
    
##  Third Party Packages

- [protobuf](https://github.com/golang/protobuf)   


