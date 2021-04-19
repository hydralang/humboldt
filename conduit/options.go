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
	"syscall"
	"time"
)

// iDialer is an interface matching that provided by net.Dialer.  This
// is the type returned by the mkDialer function.
type iDialer interface {
	// Dial connects to the address on the named network.
	Dial(network, address string) (net.Conn, error)

	// DialContext connects to the address on the named network
	// using the provided context.
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// DialerOption is an interface for options that construct a
// net.Dialer.
type DialerOption interface {
	// DialApply applies the option to a net.Dialer.
	DialApply(d *net.Dialer)
}

// dialerFilter is a filter that may be applied to DialerOption
// instances.  This is used by mechanisms supporting the LocalAddr
// option.
type dialerFilter interface {
	// DialFilter filters the option.
	DialFilter(o DialerOption) error
}

// mkDialer constructs a net.Dialer from the specified options and
// returns it.  The return type is an iDialer, allowing for testing in
// isolation.
func mkDialer(opts []DialerOption, filt dialerFilter) (iDialer, error) {
	result := &net.Dialer{}

	// Apply options
	for _, opt := range opts {
		if filt != nil {
			if err := filt.DialFilter(opt); err != nil {
				return nil, err
			}
		}
		opt.DialApply(result)
	}

	return result, nil
}

// iListenConfig is an interface matching that provided by
// net.ListenConfig.  This is the type returned by the mkListenConfig
// function.
type iListenConfig interface {
	// Listen announces on the local network address.
	Listen(ctx context.Context, network, address string) (net.Listener, error)

	// ListenPacket announces on the local network address.
	ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error)
}

// ListenerOption is an interface for options that construct a
// net.ListenConfig.
type ListenerOption interface {
	// ListenApply applies the option to a net.ListenConfig.
	ListenApply(lc *net.ListenConfig)
}

// listenerFilter is a filter that may be applied to ListenerOption
// instances.  This is used by mechanisms supporting the LocalAddr
// option.
type listenerFilter interface {
	// ListenFilter filters the option.
	ListenFilter(o ListenerOption) error
}

// mkListenConfig constructs a net.ListenConfig from the specified
// options and returns it.  The return type is an iListenConfig,
// allowing for testing in isolation.
func mkListenConfig(opts []ListenerOption, filt listenerFilter) (iListenConfig, error) {
	result := &net.ListenConfig{}

	// Apply options
	for _, opt := range opts {
		if filt != nil {
			if err := filt.ListenFilter(opt); err != nil {
				return nil, err
			}
		}
		opt.ListenApply(result)
	}

	return result, nil
}

// KeepAlive is an option for Dial and Listen that sets the KeepAlive
// option.
type KeepAlive time.Duration

// DialApply applies the option to a net.Dialer.
func (ka KeepAlive) DialApply(d *net.Dialer) {
	d.KeepAlive = time.Duration(ka)
}

// ListenApply applies the option to a net.ListenConfig.
func (ka KeepAlive) ListenApply(lc *net.ListenConfig) {
	lc.KeepAlive = time.Duration(ka)
}

// control is an option for Dial and Listen that sets the Control
// option.
type control struct {
	Control func(network, address string, c syscall.RawConn) error
}

// DialApply applies the option to a net.Dialer.
func (c control) DialApply(d *net.Dialer) {
	d.Control = c.Control
}

// ListenApply applies the option to a net.ListenConfig.
func (c control) ListenApply(lc *net.ListenConfig) {
	lc.Control = c.Control
}

// LocalAddrOption is an option that sets the LocalAddr configuration
// value for a Dialer.  It requires validation by the mechanism, so
// may not be used by some mechanisms.
type LocalAddrOption struct {
	URI  *URI     // The URI for the local side of the connection
	Addr net.Addr // The address
}

// LocalAddr returns an option that sets the LocalAddr configuration
// value for the Dialer.  It is validated by the mechanism, and may be
// unsupported by some mechanisms; mechanisms that do not support
// LocalAddr should ignore the option.
func LocalAddr(uri *URI) *LocalAddrOption {
	return &LocalAddrOption{
		URI: uri,
	}
}

// DialApply applies the option to a net.Dialer.
func (la *LocalAddrOption) DialApply(d *net.Dialer) {
	d.LocalAddr = la.Addr
}
