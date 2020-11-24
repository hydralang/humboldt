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

import "net"

// Transport describes a conduit transport mechanism.
type Transport interface {
	// Dial opens a transport in active mode; that is, for
	// connection-oriented transports, Dial causes initiation of a
	// connection.  For those transports that are not
	// connection-oriented, the conduit will still be in the
	// appropriate state.
	Dial(config interface{}, u *URI) (net.Conn, error)

	// Listen opens a transport in passive mode; that is, for
	// connection-oriented transports, Listen creates a listener
	// that may accept connections.  For those transports that are
	// not connection-oriented, the listener synthesizes the
	// appropriate events.
	Listen(config interface{}, u *URI) (net.Listener, error)
}

// transMechs is a registry of transport mechanisms.
var transMechs = map[string]Transport{}

// RegisterTransport registers a transport mechanism.
func RegisterTransport(name string, mech Transport) {
	transMechs[name] = mech
}
