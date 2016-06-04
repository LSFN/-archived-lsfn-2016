package environment

import (
	"net"
	"time"

	"github.com/LSFN/lsfn/vessel/protobuf"
)

const (
	PROTOCOL_VERSION = "0.0.2"
	NET_RATE         = 60 // packets per second
	NET_PERIOD       = ((1000 / NET_RATE) * time.Millisecond)
)

type Environment struct {
	conn           *conn
	outboundHolder *protobuf.VesselToEnvironment
	inboundHolder  *protobuf.EnvironmentToVessel
}

func NewEnvironment(environmentUDPAddress *net.UDPAddr) (*Environment, error) {
	environment := new(Environment)
	conn, err := connectToEnvironment(environmentUDPAddress)
	environment.conn = conn
	if err != nil {
		return nil, err
	}
	environment.inboundHolder = new(protobuf.EnvironmentToVessel)
	environment.outboundHolder = new(protobuf.VesselToEnvironment)
	stopChan := make(chan bool)
	go environment.send(stopChan)
	go environment.receive(stopChan)
	return environment, nil
}

func (environment *Environment) receive(stopChan chan<- bool) {
	for receivedPacket := range environment.conn.inbound {
		// Version check
		if receivedPacket.ProtocolVersion != PROTOCOL_VERSION {
			break
		}

		// Join
		environment.inboundHolder.JoinStatus = receivedPacket.JoinStatus
		if receivedPacket.JoinStatus {
			if receivedPacket.VesselID != "" {
				environment.inboundHolder.VesselID = receivedPacket.VesselID
				environment.outboundHolder.VesselID = receivedPacket.VesselID
			}
		} else {
			break
		}

		// Merge the content of the received packet with the inbound holder packet
		if receivedPacket.ShipInput != nil {
			environment.inboundHolder.ShipInput = receivedPacket.ShipInput
		}
		if receivedPacket.ShipSensors != nil {
			environment.inboundHolder.ShipSensors = receivedPacket.ShipSensors
		}
	}
	stopChan <- true
}

func (environment *Environment) send(stopChan <-chan bool) {
	ticker := time.NewTicker(NET_PERIOD)
	for {
		select {
		case <-ticker.C:
			environment.conn.outbound <- environment.outboundHolder
		case <-stopChan:
			break
		}
	}
}

func (environment *Environment) SetShipInput(input *protobuf.ShipInput) {
	environment.outboundHolder.ShipInput = input
}

func (environment *Environment) GetShipInput() *protobuf.ShipInput {
	return environment.inboundHolder.ShipInput
}

func (environment *Environment) GetShipSensors() *protobuf.ShipSensors {
	return environment.inboundHolder.ShipSensors
}
