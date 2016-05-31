package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"time"

	"github.com/LSFN/lsfn/vessel/environment"
	"github.com/LSFN/lsfn/vessel/protobuf"
)

const (
	NET_RATE         = 60 // Updates per second sent to environment
	NET_PERIOD       = (1000 / NET_RATE) * time.Millisecond
	PROTOCOL_VERSION = "0.0.1"
)

func join(environmentConn *environment.Conn) (string, error) {
	ticker := time.NewTicker(NET_PERIOD)
	defer ticker.Stop()
	stop := make(chan bool)
	defer func() {
		stop <- true
	}()
	joinMessage := &protobuf.VesselToEnvironment{
		ProtocolVersion: PROTOCOL_VERSION,
		JoinRequest:     &protobuf.JoinRequest{},
	}

	// A goroutine to repeatedly send a join request
	go func() {
		for {
			select {
			case <-ticker.C:
				environmentConn.Outbound <- joinMessage
			case <-stop:
				break
			}
		}
	}()

	for msg := range environmentConn.Inbound {
		// Protocol version check
		if msg.ProtocolVersion == "" {
			return "", errors.New("Protocol version missing from packet received from environment")
		} else if msg.ProtocolVersion != PROTOCOL_VERSION {
			return "", errors.New("Protocol version mismatch. Ours: \"" + PROTOCOL_VERSION + "\", theirs: \"" + msg.ProtocolVersion + "\"")
		}

		joinResponse := msg.GetJoinResponse()
		if joinResponse != nil {
			if joinResponse.JoinSuccessful {
				return joinResponse.VesselID, nil
			} else {
				break
			}
		}
	}
	return "", errors.New("Join unsuccessful")
}

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

	environmentConn, err := environment.ConnectToEnvironment(environmentUDPAddress)

	if err != nil {
		log.Fatalln(err)
	}

	id, err := join(environmentConn)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Joined with ID " + id)
}
