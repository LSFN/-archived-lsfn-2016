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

type shipInputStateStore struct {
	state *protobuf.ShipInput
	sync  syncNumber
}

type shipSensorsStateStore struct {
	state *protobuf.ShipSensors
	sync  syncNumber
}

type environmentStateStore struct {
	shipInput   *shipInputStateStore
	shipSensors *shipSensorsStateStore
}

type Environment struct {
	conn           *conn
	outboundHolder *protobuf.VesselToEnvironment
	stateStore     *environmentStateStore
}

func NewEnvironment(environmentUDPAddress *net.UDPAddr) (*Environment, error) {
	environment := new(Environment)

	conn, err := connectToEnvironment(environmentUDPAddress)
	environment.conn = conn
	if err != nil {
		return nil, err
	}

	environment.outboundHolder = new(protobuf.VesselToEnvironment)

	environment.stateStore = new(environmentStateStore)
	environment.stateStore.shipInput = new(shipInputStateStore)
	environment.stateStore.shipInput.state = new(protobuf.ShipInput)
	environment.stateStore.shipInput.sync = 0
	environment.stateStore.shipSensors = new(shipSensorsStateStore)
	environment.stateStore.shipSensors.state = new(protobuf.ShipSensors)
	environment.stateStore.shipSensors.sync = 0

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
		if receivedPacket.JoinStatus {
			if receivedPacket.VesselID != "" {
				environment.outboundHolder.VesselID = receivedPacket.VesselID
			}
		} else {
			break
		}

		// Merge the content of the received packet with the inbound holder packet
		var sync syncNumber = syncNumber(receivedPacket.SyncNumber)
		if receivedPacket.ShipInput != nil && sync.newerThan(environment.stateStore.shipInput.sync) {
			environment.stateStore.shipInput.state = receivedPacket.ShipInput
			environment.stateStore.shipInput.sync = sync
		}
		if receivedPacket.ShipSensors != nil && sync.newerThan(environment.stateStore.shipSensors.sync) {
			environment.stateStore.shipSensors.state = receivedPacket.ShipSensors
			environment.stateStore.shipSensors.sync = sync
		}
	}
	stopChan <- true
}

func (environment *Environment) send(stopChan <-chan bool) {
	ticker := time.NewTicker(NET_PERIOD)
	var sync syncNumber = 0
	for {
		select {
		case <-ticker.C:
			environment.conn.outbound <- environment.outboundHolder
			environment.outboundHolder.SyncNumber = uint32(sync)
			sync = sync.next()
		case <-stopChan:
			break
		}
	}
}

func (environment *Environment) SetShipInput(input *protobuf.ShipInput) {
	environment.outboundHolder.ShipInput = input
}

func (environment *Environment) GetShipInput() *protobuf.ShipInput {
	return environment.stateStore.shipInput.state
}

func (environment *Environment) GetShipSensors() *protobuf.ShipSensors {
	return environment.stateStore.shipSensors.state
}
