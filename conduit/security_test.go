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

import (
	"net"
	"testing"

	"github.com/klmitch/patcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSecurity struct {
	mock.Mock
}

func (m *mockSecurity) Dial(config interface{}, u *URI, xport Transport) (Conn, error) {
	args := m.MethodCalled("Dial", config, u, xport)

	if tmp := args.Get(0); tmp != nil {
		return tmp.(Conn), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *mockSecurity) Listen(config interface{}, u *URI, xport Transport) (net.Listener, error) {
	args := m.MethodCalled("Listen", config, u, xport)

	if tmp := args.Get(0); tmp != nil {
		return tmp.(net.Listener), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestRegisterSecurity(t *testing.T) {
	mech := &mockSecurity{}
	defer patcher.SetVar(&secMechs, map[string]Security{}).Install().Restore()

	RegisterSecurity("test", mech)

	assert.Equal(t, map[string]Security{
		"test": mech,
	}, secMechs)
}
