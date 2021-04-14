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
	"net"
	"net/url"
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

func TestTCPMechImplementsMechanism(t *testing.T) {
	assert.Implements(t, (*Mechanism)(nil), TCPMech(0))
}

func TestTCPMechDialBase(t *testing.T) {
	cfg := &mock.Mock{}
	conn := &mockConn{}
	addr := &mockAddr{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:4321",
		},
	}
	obj := TCPMech(0)
	addr.On("String").Return("127.0.0.1:1234")
	conn.On("LocalAddr").Return(addr)
	defer patcher.SetVar(&netDial, func(network, address string) (net.Conn, error) {
		assert.Equal(t, "tcp", network)
		assert.Equal(t, "127.0.0.1:4321", address)
		return conn, nil
	}).Install().Restore()

	result, err := obj.Dial(cfg, u)

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
}

func TestTCPMechDialError(t *testing.T) {
	cfg := &mock.Mock{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:4321",
		},
	}
	obj := TCPMech(0)
	defer patcher.SetVar(&netDial, func(network, address string) (net.Conn, error) {
		assert.Equal(t, "tcp", network)
		assert.Equal(t, "127.0.0.1:4321", address)
		return nil, assert.AnError
	}).Install().Restore()

	result, err := obj.Dial(cfg, u)

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
}

func TestTCPMechListenBase(t *testing.T) {
	cfg := &mock.Mock{}
	l := &mockNetListener{}
	addr := &mockAddr{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
	}
	obj := TCPMech(0)
	addr.On("String").Return("127.0.0.1:1234")
	l.On("Addr").Return(addr)
	defer patcher.SetVar(&netListen, func(network, address string) (net.Listener, error) {
		assert.Equal(t, "tcp", network)
		assert.Equal(t, "127.0.0.1:1234", address)
		return l, nil
	}).Install().Restore()

	result, err := obj.Listen(cfg, u)

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
}

func TestTCPMechListenError(t *testing.T) {
	cfg := &mock.Mock{}
	u := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
	}
	obj := TCPMech(0)
	defer patcher.SetVar(&netListen, func(network, address string) (net.Listener, error) {
		assert.Equal(t, "tcp", network)
		assert.Equal(t, "127.0.0.1:1234", address)
		return nil, assert.AnError
	}).Install().Restore()

	result, err := obj.Listen(cfg, u)

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
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
