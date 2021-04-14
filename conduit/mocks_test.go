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
	"time"

	"github.com/stretchr/testify/mock"
)

type mockAddr struct {
	mock.Mock
}

func (m *mockAddr) Network() string {
	args := m.MethodCalled("Network")

	return args.String(0)
}

func (m *mockAddr) String() string {
	args := m.MethodCalled("String")

	return args.String(0)
}

type mockConn struct {
	mock.Mock
}

func (m *mockConn) Read(b []byte) (int, error) {
	args := m.MethodCalled("Read", b)

	if tmp := args.Get(0); tmp != nil {
		data := tmp.([]byte)
		return copy(b, data), args.Error(1)
	}

	return 0, args.Error(1)
}

func (m *mockConn) Write(b []byte) (int, error) {
	args := m.MethodCalled("Write", b)

	return args.Int(0), args.Error(1)
}

func (m *mockConn) Close() error {
	args := m.MethodCalled("Close")

	return args.Error(0)
}

func (m *mockConn) LocalAddr() net.Addr {
	args := m.MethodCalled("LocalAddr")

	if tmp := args.Get(0); tmp != nil {
		return tmp.(net.Addr)
	}

	return nil
}

func (m *mockConn) RemoteAddr() net.Addr {
	args := m.MethodCalled("RemoteAddr")

	if tmp := args.Get(0); tmp != nil {
		return tmp.(net.Addr)
	}

	return nil
}

func (m *mockConn) SetDeadline(t time.Time) error {
	args := m.MethodCalled("SetDeadline", t)

	return args.Error(0)
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	args := m.MethodCalled("SetReadDeadline", t)

	return args.Error(0)
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	args := m.MethodCalled("SetWriteDeadline", t)

	return args.Error(0)
}

type mockNetListener struct {
	mock.Mock
}

func (m *mockNetListener) Accept() (net.Conn, error) {
	args := m.MethodCalled("Accept")

	if tmp := args.Get(0); tmp != nil {
		return tmp.(net.Conn), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *mockNetListener) Close() error {
	args := m.MethodCalled("Close")

	return args.Error(0)
}

func (m *mockNetListener) Addr() net.Addr {
	args := m.MethodCalled("Addr")

	if tmp := args.Get(0); tmp != nil {
		return tmp.(net.Addr)
	}

	return nil
}
