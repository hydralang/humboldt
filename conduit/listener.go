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

// Listener is a variation on the net.Listener interface that returns
// Conduit objects instead of net.Conn objects.
type Listener interface {
	// Accept waits for and returns the next conduit to the
	// listener.
	Accept() (*Conduit, error)

	// Close closes the listener.  Any blocked Accept operations
	// will be unblocked and return errors.
	Close() error

	// Addr returns the listener's network URI.
	Addr() *URI
}
