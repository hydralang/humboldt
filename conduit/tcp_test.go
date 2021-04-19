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
	"testing"

	"github.com/klmitch/patcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTCPAddr2URI(t *testing.T) {
	addr := &mockAddr{}
	addr.On("String").Return("127.0.0.1:1234")

	result := TCPAddr2URI(addr)

	assert.Equal(t, &URI{
		URL: url.URL{
			Scheme: "tcp",
			Host:   "127.0.0.1:1234",
		},
		Transport: "tcp",
	}, result)
	addr.AssertExpectations(t)
}

type mockRawConn struct {
	mock.Mock
}

func (m *mockRawConn) Control(f func(fd uintptr)) error {
	args := m.MethodCalled("Control", f)

	return args.Error(0)
}

func (m *mockRawConn) Read(f func(fd uintptr) (done bool)) error {
	args := m.MethodCalled("Read", f)

	return args.Error(0)
}

func (m *mockRawConn) Write(f func(fd uintptr) (done bool)) error {
	args := m.MethodCalled("Write", f)

	return args.Error(0)
}

func TestTCPReuseAddr(t *testing.T) {
	c := &mockRawConn{}
	c.On("Control", mock.Anything).Return(assert.AnError)
	setsockoptCalled := false
	defer patcher.SetVar(&setsockoptInt, func(fd, level, opt, value int) error {
		assert.Equal(t, 5, fd)
		assert.Equal(t, syscall.SOL_SOCKET, level)
		assert.Equal(t, syscall.SO_REUSEADDR, opt)
		assert.Equal(t, 1, value)
		setsockoptCalled = true
		return nil
	}).Install().Restore()

	err := tcpReuseAddr("net", "addr", c)

	assert.Same(t, assert.AnError, err)
	c.AssertExpectations(t)
	assert.False(t, setsockoptCalled)
	f := c.Calls[0].Arguments[0].(func(fd uintptr))
	f(uintptr(5))
	assert.True(t, setsockoptCalled)
}

func TestTCPFilterImplementsDialerFilter(t *testing.T) {
	assert.Implements(t, (*dialerFilter)(nil), tcpFilter(0))
}

func TestTCPFilterDialFilterBase(t *testing.T) {
	opt := &mockDialerOption{}
	obj := tcpFilter(0)

	err := obj.DialFilter(opt)

	assert.NoError(t, err)
	opt.AssertExpectations(t)
}

var loopback = net.IPv4(127, 0, 0, 1)

func TestTCPFilterDialFilterLocalAddr(t *testing.T) {
	opt := &LocalAddrOption{
		URI: &URI{
			URL: url.URL{
				Scheme: "tcp",
				Host:   "127.0.0.1:1234",
			},
			Transport: "tcp",
		},
	}
	obj := tcpFilter(0)

	err := obj.DialFilter(opt)

	assert.NoError(t, err)
	assert.Equal(t, &net.TCPAddr{
		IP: loopback,
	}, opt.Addr)
}

func TestTCPFilterDialFilterLocalAddrNonCanonical(t *testing.T) {
	opt := &LocalAddrOption{
		URI: &URI{
			URL: url.URL{
				Scheme: "tcp+srv",
				Host:   "127.0.0.1:1234",
			},
			Transport: "tcp",
			Discovery: "srv",
		},
	}
	obj := tcpFilter(0)

	err := obj.DialFilter(opt)

	assert.ErrorIs(t, err, ErrNotCanonical)
	assert.Nil(t, opt.Addr)
}

func TestTCPFilterDialFilterLocalAddrBadTransport(t *testing.T) {
	opt := &LocalAddrOption{
		URI: &URI{
			URL: url.URL{
				Scheme: "unix",
				Host:   "127.0.0.1:1234",
			},
			Transport: "unix",
		},
	}
	obj := tcpFilter(0)

	err := obj.DialFilter(opt)

	assert.ErrorIs(t, err, ErrUnknownTransport)
	assert.Nil(t, opt.Addr)
}

func TestTCPFilterDialFilterLocalAddrBadAddress(t *testing.T) {
	opt := &LocalAddrOption{
		URI: &URI{
			URL: url.URL{
				Scheme: "tcp",
				Host:   "127.0.0.1:1234",
			},
			Transport: "tcp",
		},
	}
	obj := tcpFilter(0)
	defer patcher.SetVar(&resolveTCPAddr, func(network, address string) (*net.TCPAddr, error) {
		return nil, assert.AnError
	}).Install().Restore()

	err := obj.DialFilter(opt)

	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, opt.Addr)
}

func TestTCPMechImplementsMechanism(t *testing.T) {
	assert.Implements(t, (*Mechanism)(nil), TCPMech(0))
}

func TestTCPMechDialBase(t *testing.T) {
	ctx := context.Background()
	cfg := &mockConfig{}
	opt := &mockDialerOption{}
	conn := &mockConn{}
	addr := &mockAddr{}
	dialer := &mockDialer{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:4321",
		},
	}
	obj := TCPMech(0)
	opt.On("DialApply", mock.Anything)
	addr.On("String").Return("127.0.0.1:1234")
	conn.On("LocalAddr").Return(addr)
	dialer.On("DialContext", ctx, "tcp", "127.0.0.1:4321").Return(conn, nil)
	defer patcher.SetVar(&mkDialerPatch, func(opts []DialerOption, filt dialerFilter) (iDialer, error) {
		assert.Equal(t, []DialerOption{opt}, opts)
		assert.Equal(t, tcpFilter(0), filt)
		return dialer, nil
	}).Install().Restore()

	result, err := obj.Dial(ctx, cfg, u, []DialerOption{opt})

	assert.NoError(t, err)
	assert.Equal(t, &Conduit{
		State: Active,
		LocalURI: &URI{
			URL: url.URL{
				Scheme: "tcp",
				Host:   "127.0.0.1:1234",
			},
			Transport: "tcp",
		},
		RemoteURI: u,
		Link:      conn,
	}, result)
	addr.AssertExpectations(t)
	conn.AssertExpectations(t)
	dialer.AssertExpectations(t)
}

func TestTCPMechDialMkDialerError(t *testing.T) {
	ctx := context.Background()
	cfg := &mockConfig{}
	opt := &mockDialerOption{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:4321",
		},
	}
	obj := TCPMech(0)
	opt.On("DialApply", mock.Anything)
	defer patcher.SetVar(&mkDialerPatch, func(opts []DialerOption, filt dialerFilter) (iDialer, error) {
		assert.Equal(t, []DialerOption{opt}, opts)
		assert.Equal(t, tcpFilter(0), filt)
		return nil, assert.AnError
	}).Install().Restore()

	result, err := obj.Dial(ctx, cfg, u, []DialerOption{opt})

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
}

func TestTCPMechDialError(t *testing.T) {
	ctx := context.Background()
	cfg := &mockConfig{}
	opt := &mockDialerOption{}
	dialer := &mockDialer{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:4321",
		},
	}
	obj := TCPMech(0)
	dialer.On("DialContext", ctx, "tcp", "127.0.0.1:4321").Return(nil, assert.AnError)
	defer patcher.SetVar(&mkDialerPatch, func(opts []DialerOption, filt dialerFilter) (iDialer, error) {
		assert.Equal(t, []DialerOption{opt}, opts)
		assert.Equal(t, tcpFilter(0), filt)
		return dialer, nil
	}).Install().Restore()

	result, err := obj.Dial(ctx, cfg, u, []DialerOption{opt})

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
	dialer.AssertExpectations(t)
}

func TestTCPMechListenBase(t *testing.T) {
	ctx := context.Background()
	cfg := &mockConfig{}
	opt := &mockListenerOption{}
	l := &mockNetListener{}
	addr := &mockAddr{}
	lc := &mockListenConfig{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
	}
	obj := TCPMech(0)
	addr.On("String").Return("127.0.0.1:1234")
	l.On("Addr").Return(addr)
	lc.On("Listen", ctx, "tcp", "127.0.0.1:1234").Return(l, nil)
	defer patcher.SetVar(&mkListenConfigPatch, func(opts []ListenerOption, filt listenerFilter) (iListenConfig, error) {
		assert.Len(t, opts, 2)
		assert.Equal(t, opt, opts[0])
		assert.Nil(t, filt)
		return lc, nil
	}).Install().Restore()

	result, err := obj.Listen(ctx, cfg, u, []ListenerOption{opt})

	assert.NoError(t, err)
	assert.Equal(t, &TCPListener{
		L: l,
		URI: &URI{
			URL: url.URL{
				Scheme: "tcp",
				Host:   "127.0.0.1:1234",
			},
			Transport: "tcp",
		},
	}, result)
	addr.AssertExpectations(t)
	l.AssertExpectations(t)
	lc.AssertExpectations(t)
}

func TestTCPMechListenMkListenConfigError(t *testing.T) {
	ctx := context.Background()
	cfg := &mockConfig{}
	opt := &mockListenerOption{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
	}
	obj := TCPMech(0)
	defer patcher.SetVar(&mkListenConfigPatch, func(opts []ListenerOption, filt listenerFilter) (iListenConfig, error) {
		assert.Len(t, opts, 2)
		assert.Equal(t, opt, opts[0])
		assert.Nil(t, filt)
		return nil, assert.AnError
	}).Install().Restore()

	result, err := obj.Listen(ctx, cfg, u, []ListenerOption{opt})

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
}

func TestTCPMechListenError(t *testing.T) {
	ctx := context.Background()
	cfg := &mockConfig{}
	opt := &mockListenerOption{}
	lc := &mockListenConfig{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
	}
	obj := TCPMech(0)
	lc.On("Listen", ctx, "tcp", "127.0.0.1:1234").Return(nil, assert.AnError)
	defer patcher.SetVar(&mkListenConfigPatch, func(opts []ListenerOption, filt listenerFilter) (iListenConfig, error) {
		assert.Len(t, opts, 2)
		assert.Equal(t, opt, opts[0])
		assert.Nil(t, filt)
		return lc, nil
	}).Install().Restore()

	result, err := obj.Listen(ctx, cfg, u, []ListenerOption{opt})

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
	lc.AssertExpectations(t)
}

func TestTCPListenerImplementsListener(t *testing.T) {
	assert.Implements(t, (*Listener)(nil), &TCPListener{})
}

func TestTCPListenerAcceptBase(t *testing.T) {
	addr := &mockAddr{}
	c := &mockConn{}
	l := &mockNetListener{}
	obj := &TCPListener{
		L: l,
		URI: &URI{
			URL: url.URL{
				Host: "127.0.0.1:4321",
			},
		},
	}
	addr.On("String").Return("127.0.0.1:1234")
	c.On("RemoteAddr").Return(addr)
	l.On("Accept").Return(c, nil)

	result, err := obj.Accept()

	assert.NoError(t, err)
	assert.Equal(t, &Conduit{
		State:    Passive,
		LocalURI: obj.URI,
		RemoteURI: &URI{
			URL: url.URL{
				Scheme: "tcp",
				Host:   "127.0.0.1:1234",
			},
			Transport: "tcp",
		},
		Link: c,
	}, result)
	l.AssertExpectations(t)
	c.AssertExpectations(t)
	addr.AssertExpectations(t)
}

func TestTCPListenerAcceptError(t *testing.T) {
	l := &mockNetListener{}
	obj := &TCPListener{
		L: l,
		URI: &URI{
			URL: url.URL{
				Host: "127.0.0.1:4321",
			},
		},
	}
	l.On("Accept").Return(nil, assert.AnError)

	result, err := obj.Accept()

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
	l.AssertExpectations(t)
}

func TestTCPListenerClose(t *testing.T) {
	l := &mockNetListener{}
	obj := &TCPListener{
		L: l,
	}
	l.On("Close").Return(assert.AnError)

	err := obj.Close()

	assert.Same(t, assert.AnError, err)
	l.AssertExpectations(t)
}

func TestTCPListenerAddr(t *testing.T) {
	obj := &TCPListener{
		URI: &URI{
			URL: url.URL{
				Host: "127.0.0.1:4321",
			},
		},
	}

	result := obj.Addr()

	assert.Equal(t, &URI{
		URL: url.URL{
			Host: "127.0.0.1:4321",
		},
	}, result)
}
