// Copyright (c) 2021 Kevin L. Mitchell
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

import "net"

// State indicates the state the conduit is in.
type State int

// Defined conduit states.
const (
	Undefined State = iota // Conduit is in an undefined state
	Active                 // Conduit is new outgoing conduit
	Passive                // Conduit is new incoming conduit
	Open                   // Conduit is fully established and open
	Closed                 // Conduit has been closed
	Error                  // Conduit has received an error
)

// Conduit describes an established conduit.
type Conduit struct {
	State        State       // The state the conduit is in
	Error        error       // When in Error state, this contains the error
	MinProto     uint32      // Minimum supported protocol version
	MaxProto     uint32      // Maximum supported protocol version
	Proto        uint32      // Selected protocol version
	RTT          uint32      // Estimated round-trip time
	Deviation    uint32      // Estimated round-trip time deviation
	Peer         interface{} // Peer or client description
	Confidential bool        // Flag indicating conduit is confidential
	Integrity    bool        // Flag indicating conduit is integrity-protected
	Principal    string      // Name of the principal from security layer
	Strength     uint32      // Estimate of the encryption strength
	LocalURI     *URI        // Local conduit URI
	RemoteURI    *URI        // Remote conduit URI
	Link         net.Conn    // Network connection
}
