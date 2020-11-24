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
