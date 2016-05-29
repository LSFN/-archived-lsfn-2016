package main

import (
	"flag"
	"net"

	"github.com/Lukeus_Maximus/lsfn/vessel/environment"
)

func main() {
	var environmentIPStr string
	var environmentPort int

	flag.StringVar(&environmentIPStr, "ip", "127.0.0.1", "The IP address of the LSFN environment server to connect to.")
	flag.IntVar(&environmentPort, "port", 39461, "The port of the LSFN environment server to connect to.")

	flag.Parse()

	environmentIP := net.ParseIP(environmentIPStr)

	environmentUDPAddress := &net.UDPAddr{
		IP:   environmentIP,
		Port: environmentPort,
	}

	_, err := environment.ConnectToEnvironment(environmentUDPAddress)

	if err != nil {
		return
	}

}
