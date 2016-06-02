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

Every packet is sent with a string-based protocol version.

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
by the environment and are present on every packet that is sent from the vessel.

#### Game Join Negotiation ####

A vessel is either joining or rejoining the environment. If it is rejoining then it
has been previously issued with a vessel ID and it should send this with every
packet it sends to the environment. If the vessel hasn't been issued one then it
can simply be omitted from the packets the vessel sends.

When receiving packets from a vessel, the environment inspects the vessel ID. If it
recognises the vessel ID then it sets the join status boolean to true on all future
packets it sends to the vessel. If it doesn't or there is no vessel ID then it
decides whether to let the vessel join the game. If it will allow the join, it sets
the join status boolean to true on future packets and sends a new vessel ID. If it
denies the join, the join status boolean is set to false.

If the vessel receives a join status of false then it must stop any further
communication with the vessel. If the join status is true and the environment has
not provided a vessel ID, the vessel assumes that it has rejoined successfully.
If the join status is true and the environment is providing a vessel ID, the vessel
has joined successfully and sets the vessel ID on every future packet it sends. In 
the case where the a rejoining vessel receives a different vessel ID back from the
environment, the vessel has joined successfully and must set the new vessel ID on
every future packet but it must also erase any previously stored state received
from the environment. 

###  ###


