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

// link is a type internal to the conduit library.  It is an interface
// that is used for sending messages to the peer or for closing the
// conduit.
type link interface {
	// Send sends a message on the conduit to the peer.
	Send(msg []byte) error

	// Close shuts down the conduit.
	Close()
}
