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

import "context"

// Discovery describes a discovery mechanism.
type Discovery interface {
	// Discover is passed a URI and returns a list of canonical
	// URIs retrieved from the discovery mechanism.  The list may
	// be in a priority order, or may be in an arbitrary
	// randomized order, depending on the mechanism.
	Discover(u *URI) ([]*URI, error)
}

// discMechs is a registry of discovery mechanisms.
var discMechs = map[string]Discovery{}

// RegisterDiscovery registers a discovery mechanism.
func RegisterDiscovery(name string, mech Discovery) {
	discMechs[name] = mech
}

// LookupDiscovery is used to look up a discovery mechanism.
func LookupDiscovery(name string) Discovery {
	return discMechs[name]
}

// Mechanism describes a conduit transport or security mechanism.
type Mechanism interface {
	// Dial opens a conduit in active mode; that is, for
	// connection-oriented transports, Dial causes initiation of a
	// connection.  For those transports that are not
	// connection-oriented, the conduit will still be in the
	// appropriate state.
	Dial(ctx context.Context, config Config, u *URI, opts []DialerOption) (*Conduit, error)

	// Listen opens a transport in passive mode; that is, for
	// connection-oriented transports, Listen creates a listener
	// that may accept connections.  For those transports that are
	// not connection-oriented, the listener synthesizes the
	// appropriate state.
	Listen(ctx context.Context, config Config, u *URI, opts []ListenerOption) (Listener, error)
}

// secMechs is a registry of security layer mechanisms.
var secMechs = map[string]Mechanism{}

// RegisterSecurity registers a security layer mechanism.
func RegisterSecurity(name string, mech Mechanism) {
	secMechs[name] = mech
}

// LookupSecurity is used to look up a security layer mechanism.
func LookupSecurity(name string) Mechanism {
	return secMechs[name]
}

// transMechs is a registry of transport mechanisms.
var transMechs = map[string]Mechanism{}

// RegisterTransport registers a transport mechanism.
func RegisterTransport(name string, mech Mechanism) {
	transMechs[name] = mech
}

// LookupTransport is used to look up a transport mechanism.
func LookupTransport(name string) Mechanism {
	return transMechs[name]
}
