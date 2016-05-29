package environment

//go:generate protoc3 -I $GOPATH/src/github.com/LSFN/lsfn/protobuf --go_out=../protobuf $GOPATH/src/github.com/LSFN/lsfn/protobuf/environmentToVessel.proto $GOPATH/src/github.com/LSFN/lsfn/protobuf/vesselToEnvironment.proto

import (
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/LSFN/lsfn/vessel/protobuf"
)

const (
	MESSAGE_BUFFER_SIZE = 10
)

type Conn struct {
	inbound  <-chan *protobuf.EnvironmentToVessel
	outbound chan<- *protobuf.VesselToEnvironment
}

func ConnectToEnvironment(environmentUDPAddress *net.UDPAddr) (*Conn, error) {
	conn, err := net.DialUDP("udp", nil, environmentUDPAddress)
	if err != nil {
		return nil, err
	}
	inboundMessages := make(chan *protobuf.EnvironmentToVessel, MESSAGE_BUFFER_SIZE)
	outboundMessages := make(chan *protobuf.VesselToEnvironment, MESSAGE_BUFFER_SIZE)
	environmentConnection := &Conn{
		inbound:  inboundMessages,
		outbound: outboundMessages,
	}
	go readFromServer(conn, inboundMessages)
	go writeToServer(conn, outboundMessages)
	return environmentConnection, nil
}

func readFromServer(conn *net.UDPConn, inboundMessages chan<- *protobuf.EnvironmentToVessel) {
	var readBuf []byte
	for {
		_, _, err := conn.ReadFromUDP(readBuf)
		if err != nil {
			break
		}
		msg := &protobuf.EnvironmentToVessel{}
		err = proto.Unmarshal(readBuf, msg)
		if err != nil {
			break
		}
		inboundMessages <- msg
	}
}

func writeToServer(conn *net.UDPConn, outboundMessages <-chan *protobuf.VesselToEnvironment) {
	for msg := range outboundMessages {
		buf, err := proto.Marshal(msg)
		if err != nil {
			break
		}
		_, err = conn.Write(buf)
		if err != nil {
			break
		}
	}
}
