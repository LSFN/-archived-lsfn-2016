# Vessel - Environment Protocol #

## Overview ##

This protocol is designed for use over UDP. It aims to update state on the
receiving end by sending idempotent state updates in each packet. Packets have a
number that allows the receiver to determine the recency of the received data. If
the receiver has already received data that is marked as more recent than that
contained in a packet it has received then the receiver is expected to ignore that
packet. This prevents problems that may be caused by receiving out-of-order UDP
packets.

It is expected that packets are sent by both vessel and environment at the same
rate as the game simulation occurs to enable real-time communication.

The environment is considered to be the one true represenation of the game state.
Each vessel will send its input state to the environment; this contains both
gameplay input and other administrative input (such as game join negotiation). The
environment will send to each vessel the ship's gameplay input, the ship's sensor
data and any administrative state.

## Main Protocol Parts ##

### Protocol Version ###

Every packet is sent with a string-based protocol version in the field
`protocolVersion`.

#### Protocol Version Negotiation ####

Vessels start by sending their protocol version to the environment. The environment
will always respond with packets containing its own protocol version.

If the two match then all packets sent at that version are expected to have been
understood (but not necessarily received) by the receiving end and communication
continues unabated.

If the the environment receives a protocol version from a vessel that doesn't match
its own then it responds with packets that contain nothing but the environment's
version. It will send these packets until it stops receiving packets from the
vessel. The vessel will cease sending packets to the environment if it receives a
packet containing a protocol version different to its own.

### Vessel Identifier ###

The vessel identifier identifies the vessel to the environment. These are issued
by the environment and are present on every packet that is sent from the vessel in
the field `vesselID`.

#### Game Join Negotiation ####

A vessel is either joining or rejoining the environment. If it is rejoining then it
has been previously issued with a `vesselID` and it should send this with every
packet it sends to the environment. If the vessel hasn't been issued one then it
can simply be omitted from the packets the vessel sends.

When receiving packets from a vessel, the environment inspects `vesselID`. If
it recognises the vesselID then it sets the `joinStatus` boolean to `true` on all
future packets it sends to the vessel. If it doesn't or there is no vessel ID then
it decides whether to let the vessel join the game. If it will allow the join, it
sets `joinStatus` to `true` on future packets and sends a new vessel ID. If it
denies the join, `joinStatus` is set to `false`.

If the vessel receives a `joinStatus` of `false` then it must stop any further
communication with the vessel. If `joinStatus` is `true` and the environment has
not provided a vessel ID, the vessel assumes that it has rejoined successfully.
If `joinStatus` is `true` and the environment is providing a vessel ID, the vessel
has joined successfully and sets `vesselID` on every future packet it sends. In 
the case where the a rejoining vessel receives a different vessel ID back from the
environment, the vessel has joined successfully and must set the new vessel ID on
every future packet but it must also erase any previously stored state received
from the environment.

If the environment issued a new vessel ID, the environment should stop sending the
vessel ID to the vessel when the vessel starts sending that vessel ID on the
packets it sends to the environment.

### Sync Number ###

To prevent packets received out-of-order by the environment from causing
undesirable game state changes, each packet will be sent with information that
allows the receiver to determine if the content of the packet is the latest it
has received.

This is implemented as the one-byte field `syncNumber`. Each packet sent by either
the vessel or environment will have this field set. This number increments with
each sent packet. If the sync number is to be incremeted beyond 255, it is set
to 0. Thus, the sync numbers move in a 256-step cycle.

When information is received in a packet, it should be stored with the packet's
sync number. When packet is received with information that would overwrite the
stored version, the sync number of the stored version and the packet's version are
compared. If the packet's sync number is 128 or fewer ahead in the sync number
cycle, the packet's version is assumed to be more recent. Otherwise, it is assumed
to be out-of-date and will therefore be discarded.

#### Comparison Algorithm ####

The following psuedocode describes the sync number comparison algorithm.
```
s = stored version sync number
p = packet `syncNumber`
if s < 128 {
	if s < p < s + 128 {
		packet version is more recent
	} else {
		packet version is out-of-date
	}
} else {
	if s - 128 < p < s {
		packet version is out-of-date
	} else {
		packet version is more recent
	}
}
```

#### Packet Granularity ####

An important feature of the protocol is the ability to split updates to state into
different packets. Because the protocol uses UDP, it is limited to a packet size of
slightly less than 2^16 bytes. In the case where the information we want to send to
the receiver doesn't fit in a single packet, it can simply be divided among several
packets.

However, as UDP is unreliable, it is the case that not all of these packets will be
received resulting in only some of the updates being applied. Depending on what the
split-up information is, the consequeces of receiving partial information could be
worse than simply not receiving that information.

For example, a ship with two engines has two separate input throttles. If these
throttle inputs were split between packets then it could be the case that a player
that intended to increase both throttles simultaneously sees their ship veer in one
direction because the environment received an update to one throttle but not the
other. It would be far better for the throttle inputs not to be split between
packets to prevent undesirable ship movement.

This protocol must be designed such that information split between packets doesn't
cause chaos in this fashion.

### Ship Input State ###

The vessel controls its ship on the environment by sending the environment what it
thinks all of the control inputs to the ship should currently be set to. This
includes throttle positions, thruster switches and weapon triggers.

The environment also sends the same set of information back to the vessel. This
shows the vessel the input that the environment is actually applying to the ship.

The following fields are sent
* Engine throttles x 2 (left and right) as floating point values between 0 and 1
* Thrusters x 8 (2 at each end of each side of a square ship) as booleans
* Gun trigger x 1 as a boolean

Note: The gun trigger is as on an automatic weapon. A `true` value indicates that
the weapon should fire continuously whilst `false` indicates that it should not.

### Ship Sensor Readings ###

The vessel receives information on the ship's surroundings. The information is as
you might expect to receive from a simple visual sensor giving information from the
ship's immediate surroundings. It is the environment that determines how far away
the ship can actually sense objects. To ensure the vessel knows only what it is
meant to know, the environment only sends information about objects within that
vessel's ship's sensor range. Objects that no information is sent for might as well
not exist to the receiving vessel.

The following fields are sent *for each object*
* Object type (ship or bullet) as an enum value
* Position (x and y) as floating point values
* Orientation as a floating point value in radians
* Velocity (x and y) as floating point values
These values are given using the ship that is seeing the object as the point of
reference. Therefore, if a ship is infront of the viewing ship and facing towards
it then it would have an orientation of pi radians, an x-position of 0 and a
y-position of some positive number. To better understand the coordinate system of a
ship, imagine the ship's center to be the origin of a standard 2D coordinate space.
+X is to the right and +Y is forwards. A ship always points to +Y in its own
coordinate system.

The ship also senses one piece of information about itself which is its health.
This is a single integer field separate from the visual sensors.

