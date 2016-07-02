# Interface - Vessel Protocol #

## Overview ##

This protocol is designed for use over TCP. It aims to update state on the
receiving end by sending idempotent state updates in each packet.

It is expected that packets are sent by both interface and vessel at the same
rate as the game simulation occurs to enable real-time communication.

The vessel will receive input from multiple interfaces. The inputs received may
conflict with one another or otherwise overlap. It is the job of the vessel to
deconflict these using some sensible strategy to determine what the ship's overall
control input is. The vessel will also inform interfaces of what the control input
is currently so that it may be displayed.

## Main Protocol Parts ##

### Protocol Version ###

Upon a connection being established, the first thing that must be sent by both
interface and vessel is the protocol version. This is contained in the string field
`protocolVersion`. Both interface and vessel listen for the protocol version from
the other and check the protocol version received against their own (which they
sent). If they don't match, they disconnect one another. Otherwise they communicate
via this protocol version.

Only the first message in either direction must contain the protocol version, it
should be ignored therafter and doesn't need to be added to future messages.

### Ship Input ###

When an interface wishes to provide changes for one or more ship inputs, it sets
only the relevant fields on a `ShipInput` message and sends it to the vessel in the
`shipInput` field. The vessel updates all the parts of the combined ship input that
were specified in the received message. Any parts of the input that were
unspecified by the interface are left unaltered.

The vessel can update the interfaces with the true ship input that it has received
from the environment by sending it to the interfaces in the `shipInput` field.

### Ship Sensors ###

To update the interfaces' view of the world, the vessel can send a ship sensors
message in the `shipSensors` field.
