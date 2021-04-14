// Copyright (c) 2020, 2021 Kevin L. Mitchell
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

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// URI describes a full Humboldt conduit URI.  It is a variation on
// url.URL that breaks down the scheme.
type URI struct {
	url.URL // URL fields

	Transport string // The transport identifier
	Security  string // The security layer identifier
	Discovery string // The discovery mechanism identifier
}

// Parse parses a raw URL into a conduit URI.  It is based on
// url.Parse.
func Parse(rawuri string) (*URI, error) {
	result := &URI{}

	// Begin by parsing the URL
	tmpURL, err := url.Parse(rawuri)
	if err != nil {
		return nil, err
	}
	result.URL = *tmpURL

	// Disassemble the scheme into its component parts
	scheme := tmpURL.Scheme
	discIdx := strings.LastIndexAny(scheme, ".")
	if discIdx >= 0 {
		result.Discovery = scheme[discIdx+1:]
		scheme = scheme[:discIdx]
	}
	secIdx := strings.IndexAny(scheme, "+")
	if secIdx >= 0 {
		result.Security = scheme[secIdx+1:]
		scheme = scheme[:secIdx]
	}
	result.Transport = scheme

	return result, nil
}

// IsCanonical tests if the conduit URI is canonical.  To be
// canonical, no discovery mechanism may be specified, and the host
// must be a raw IP address and the port must be numeric.  (If there
// is no Host in the URI, the URI is canonical unless a discovery
// mechanism was specified.)
func (u *URI) IsCanonical() bool {
	// If there's a discovery mechanism, the URI is not canonical
	if u.Discovery != "" {
		return false
	}

	// If there's no host information, the URI is canonical
	if u.Host == "" {
		return true
	}

	// Split the host and port; use the net.SplitHostPort function
	// instead of URL.Hostname() and URL.Port() because it balks
	// at non-numeric ports and we want to allow those
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return false
	}

	// Now try parsing the hostname as an IP and the port as a
	// number and see what happens...
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return false
	}
	if net.ParseIP(host) == nil {
		return false
	}

	return true
}

// Canonicalize canonicalizes a conduit URI.  It returns a list of
// conduit URIs, as it will call discovery mechanisms and include all
// known IPs for a given hostname.
func (u *URI) Canonicalize() ([]*URI, error) {
	// If there's a discovery mechanism, look it up and call it
	if u.Discovery != "" {
		disc, ok := discMechs[u.Discovery]
		if !ok {
			return nil, fmt.Errorf("%q: %w", u.Discovery, ErrUnknownDiscovery)
		}

		return disc.Discover(u)
	}

	// If there's no host information, then the URI is canonical
	if u.Host == "" {
		return []*URI{u}, nil
	}

	// Split the host and port; use the net.SplitHostPort function
	// instead of URL.Hostname() and URL.Port() because it balks
	// at non-numeric ports and we want to allow those
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, err
	}

	// Build up the list of IPs
	ips := []net.IP{}
	if ip := net.ParseIP(host); ip != nil {
		ips = append(ips, ip)
	} else {
		var err error
		if ips, err = lookupIP(host); err != nil {
			return nil, err
		}
	}

	// Determine the canonical port
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		// Must be a service name
		numPort, err := lookupPort(u.Transport, port)
		if err != nil {
			return nil, err
		}
		port = strconv.FormatUint(uint64(numPort), 10)
	}

	// Assemble the list of URIs
	URIs := make([]*URI, 0, len(ips))
	for _, ip := range ips {
		URIs = append(URIs, &URI{
			URL: url.URL{
				Scheme:     u.Scheme,
				Opaque:     u.Opaque,
				User:       u.User,
				Host:       net.JoinHostPort(ip.String(), port),
				Path:       u.Path,
				RawPath:    u.RawPath,
				ForceQuery: u.ForceQuery,
				RawQuery:   u.RawQuery,
				Fragment:   u.Fragment,
			},
			Transport: u.Transport,
			Security:  u.Security,
			Discovery: u.Discovery,
		})
	}

	return URIs, nil
}

// Dial opens a conduit in active mode; that is, for
// connection-oriented transports, Dial causes initiation of a
// connection.  For those transports that are not connection-oriented,
// the conduit will still be in the appropriate state.
func (u *URI) Dial(config interface{}) (*Conduit, error) {
	if !u.IsCanonical() {
		return nil, fmt.Errorf("%s: %w", u, ErrNotCanonical)
	}

	// Make sure there's a transport mechanism
	if u.Transport == "" {
		return nil, fmt.Errorf("%s: %q: %w", u, u.Transport, ErrUnknownTransport)
	}

	// Is there a security layer?
	if u.Security != "" {
		if mech := lookupSecurity(u.Security); mech != nil {
			return mech.Dial(config, u)
		}
		return nil, fmt.Errorf("%s: %q: %w", u, u.Security, ErrUnknownSecurity)
	}

	if mech := lookupTransport(u.Transport); mech != nil {
		return mech.Dial(config, u)
	}
	return nil, fmt.Errorf("%s: %q: %w", u, u.Transport, ErrUnknownTransport)
}

// Listen opens a transport in passive mode; that is, for
// connection-oriented transports, Listen creates a listener that may
// accept connections.  For those transports that are not
// connection-oriented, the listener synthesizes the appropriate
// state.
func (u *URI) Listen(config interface{}) (Listener, error) {
	if !u.IsCanonical() {
		return nil, fmt.Errorf("%s: %w", u, ErrNotCanonical)
	}

	// Make sure there's a transport mechanism
	if u.Transport == "" {
		return nil, fmt.Errorf("%s: %q: %w", u, u.Transport, ErrUnknownTransport)
	}

	// Is there a security layer?
	if u.Security != "" {
		if mech := lookupSecurity(u.Security); mech != nil {
			return mech.Listen(config, u)
		}
		return nil, fmt.Errorf("%s: %q: %w", u, u.Security, ErrUnknownSecurity)
	}

	if mech := lookupTransport(u.Transport); mech != nil {
		return mech.Listen(config, u)
	}
	return nil, fmt.Errorf("%s: %q: %w", u, u.Transport, ErrUnknownTransport)
}

// Dial opens a conduit in active mode; that is, for
// connection-oriented transports, Dial causes initiation of a
// connection.  For those transports that are not connection-oriented,
// the conduit will still be in the appropriate state.
func Dial(config interface{}, uri string) (*Conduit, error) {
	// Parse the URI
	u, err := Parse(uri)
	if err != nil {
		return nil, err
	}

	// Dial it
	return u.Dial(config)
}

// Listen opens a transport in passive mode; that is, for
// connection-oriented transports, Listen creates a listener that may
// accept connections.  For those transports that are not
// connection-oriented, the listener synthesizes the appropriate
// state.
func Listen(config interface{}, uri string) (Listener, error) {
	// Parse the URI
	u, err := Parse(uri)
	if err != nil {
		return nil, err
	}

	// Dial it
	return u.Listen(config)
}
