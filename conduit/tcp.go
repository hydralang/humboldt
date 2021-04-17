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

import (
	"context"
	"net"
	"net/url"
	"syscall"
)

// TCPAddr2URI converts an address in the form returned by TCP
// connections into an appropriate URI.
func TCPAddr2URI(addr net.Addr) *URI {
	return &URI{
		URL: url.URL{
			Scheme: "tcp",
			Host:   addr.String(),
		},
		Transport: "tcp",
	}
}

// tcpReuseAddr is an implementation of the Control option which sets
// the "reuseaddr" flag on a listening socket.
func tcpReuseAddr(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		setsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1) //nolint:errcheck
	})
}

// TCPMech is a mechanism for TCP connections.
type TCPMech int

// Dial opens a conduit in active mode; that is, for
// connection-oriented transports, Dial causes initiation of a
// connection.  For those transports that are not connection-oriented,
// the conduit will still be in the appropriate state.
func (t TCPMech) Dial(ctx context.Context, config Config, u *URI, opts []DialerOption) (*Conduit, error) {
	// Construct the dialer
	dialer := mkDialerPatch(opts)

	// Dial the target
	c, err := dialer.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		return nil, err
	}

	// Construct and return a Conduit
	return &Conduit{
		State:     Active,
		LocalURI:  TCPAddr2URI(c.LocalAddr()),
		RemoteURI: u,
		Link:      c,
	}, nil
}

// Listen opens a transport in passive mode; that is, for
// connection-oriented transports, Listen creates a listener that may
// accept connections.  For those transports that are not
// connection-oriented, the listener synthesizes the appropriate
// state.
func (t TCPMech) Listen(ctx context.Context, config Config, u *URI, opts []ListenerOption) (Listener, error) {
	// Construct the listener config
	opts = append(opts, control{Control: tcpReuseAddr})
	lc := mkListenConfigPatch(opts)

	// Create the listener
	l, err := lc.Listen(ctx, "tcp", u.Host)
	if err != nil {
		return nil, err
	}

	// Return a listener
	return &TCPListener{
		L:   l,
		URI: TCPAddr2URI(l.Addr()),
	}, nil
}

// TCPListener is an implementation of Listener for the TCP transport.
type TCPListener struct {
	L   net.Listener // Underlying TCP listener
	URI *URI         // URI contains the URI used to open the listener
}

// Accept waits for and returns the next conduit to the listener.
func (l *TCPListener) Accept() (*Conduit, error) {
	// Accept a connection
	c, err := l.L.Accept()
	if err != nil {
		return nil, err
	}

	// Wrap it in a conduit
	return &Conduit{
		State:     Passive,
		LocalURI:  l.URI,
		RemoteURI: TCPAddr2URI(c.RemoteAddr()),
		Link:      c,
	}, nil
}

// Close closes the listener.  Any blocked Accept operations will be
// unblocked and return errors.
func (l *TCPListener) Close() error {
	return l.L.Close()
}

// Addr returns the listener's network URI.
func (l *TCPListener) Addr() *URI {
	return l.URI
}

// init initializes the TCP transport.
func init() {
	RegisterTransport("tcp", TCPMech(0))
}
