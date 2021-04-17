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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDialer struct {
	mock.Mock
}

func (m *mockDialer) Dial(network, address string) (net.Conn, error) {
	args := m.MethodCalled("Dial", network, address)

	if tmp := args.Get(0); tmp != nil {
		return tmp.(net.Conn), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *mockDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	args := m.MethodCalled("DialContext", ctx, network, address)

	if tmp := args.Get(0); tmp != nil {
		return tmp.(net.Conn), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockDialerOption struct {
	mock.Mock
}

func (m *mockDialerOption) DialApply(d *net.Dialer) {
	m.MethodCalled("DialApply", d)
}

func TestMkDialer(t *testing.T) {
	opt1 := &mockDialerOption{}
	opt2 := &mockDialerOption{}
	opt1.On("DialApply", &net.Dialer{})
	opt2.On("DialApply", &net.Dialer{}).Run(func(args mock.Arguments) {
		dialer := args[0].(*net.Dialer)
		dialer.KeepAlive = time.Hour
	})

	result := mkDialer([]DialerOption{opt1, opt2})

	assert.Equal(t, &net.Dialer{
		KeepAlive: time.Hour,
	}, result)
	opt1.AssertExpectations(t)
	opt2.AssertExpectations(t)
}

type mockListenConfig struct {
	mock.Mock
}

func (m *mockListenConfig) Listen(ctx context.Context, network, address string) (net.Listener, error) {
	args := m.MethodCalled("Listen", ctx, network, address)

	if tmp := args.Get(0); tmp != nil {
		return tmp.(net.Listener), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *mockListenConfig) ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	args := m.MethodCalled("ListenPacket", ctx, network, address)

	if tmp := args.Get(0); tmp != nil {
		return tmp.(net.PacketConn), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockListenerOption struct {
	mock.Mock
}

func (m *mockListenerOption) ListenApply(d *net.ListenConfig) {
	m.MethodCalled("ListenApply", d)
}

func TestMkListenConfig(t *testing.T) {
	opt1 := &mockListenerOption{}
	opt2 := &mockListenerOption{}
	opt1.On("ListenApply", &net.ListenConfig{})
	opt2.On("ListenApply", &net.ListenConfig{}).Run(func(args mock.Arguments) {
		dialer := args[0].(*net.ListenConfig)
		dialer.KeepAlive = time.Hour
	})

	result := mkListenConfig([]ListenerOption{opt1, opt2})

	assert.Equal(t, &net.ListenConfig{
		KeepAlive: time.Hour,
	}, result)
	opt1.AssertExpectations(t)
	opt2.AssertExpectations(t)
}

func TestKeepAliveImplementsDialerOption(t *testing.T) {
	assert.Implements(t, (*DialerOption)(nil), KeepAlive(time.Second))
}

func TestKeepAliveImplementsListenerOption(t *testing.T) {
	assert.Implements(t, (*ListenerOption)(nil), KeepAlive(time.Second))
}

func TestKeepAliveDialApply(t *testing.T) {
	dialer := &net.Dialer{}
	obj := KeepAlive(time.Second)

	obj.DialApply(dialer)

	assert.Equal(t, &net.Dialer{
		KeepAlive: time.Second,
	}, dialer)
}

func TestKeepAliveListenApply(t *testing.T) {
	lc := &net.ListenConfig{}
	obj := KeepAlive(time.Second)

	obj.ListenApply(lc)

	assert.Equal(t, &net.ListenConfig{
		KeepAlive: time.Second,
	}, lc)
}

func TestControlImplementsDialerOption(t *testing.T) {
	assert.Implements(t, (*DialerOption)(nil), control{})
}

func TestControlImplementsListenerOption(t *testing.T) {
	assert.Implements(t, (*ListenerOption)(nil), control{})
}

func TestControlDialApply(t *testing.T) {
	dialer := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return nil
		},
	}
	obj := control{}

	obj.DialApply(dialer)

	assert.Equal(t, &net.Dialer{}, dialer)
}

func TestControlListenApply(t *testing.T) {
	lc := &net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return nil
		},
	}
	obj := control{}

	obj.ListenApply(lc)

	assert.Equal(t, &net.ListenConfig{}, lc)
}
