// Copyright (c) 2020 Kevin L. Mitchell
//
// Licensed under the Apache License, Version 2.0 (the "License"); you
// may not use this file except in compliance with the License.  You
// may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.  See the License for the specific language governing
// permissions and limitations under the License.

package conduit

// Event indicates what sort of event occurred on the conduit.
type Event int

// Defined conduit events.
const (
	Received    Event = iota // Data was received from the conduit
	Closed                   // The conduit was closed
	Error                    // An error occurred
	Active                   // A new active conduit has been created
	Passive                  // A new passive conduit has been created
	DialError                // An error occurred while dialing
	ListenError              // The listener produced an error
)

// Conduit is a description of an active conduit to a client or a
// peer.
type Conduit struct {
	Confidential bool        // Flag indicating conduit is confidential
	Integrity    bool        // Flag indicating conduit is integrity-protected
	Principal    string      // Name of the principal from security layer
	Strength     uint32      // Estimate of the encryption strength
	MinProto     uint32      // Minimum protocol version
	MaxProto     uint32      // Maximum protocol version
	Proto        uint32      // Selected protocol version
	LocalURI     *URI        // Local conduit
	RemoteURI    *URI        // Remote conduit
	RTT          uint32      // Estimated round-trip time
	Deviation    uint32      // Estimated round-trip time deviation
	Peer         interface{} // Peer or client description

	link link // The actual link; used to send messages and close conduit
}

// Send sends a message over the conduit to the peer, whether a client
// or another Humboldt node.
func (c *Conduit) Send(msg []byte) error {
	return c.link.Send(msg)
}

// Close closes the conduit.  A notification will be sent to the
// protocol processor indicating that the conduit has been closed.
func (c *Conduit) Close() {
	c.link.Close()
}
