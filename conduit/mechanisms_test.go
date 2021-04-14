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
	"testing"

	"github.com/klmitch/patcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDiscovery struct {
	mock.Mock
}

func (m *mockDiscovery) Discover(u *URI) ([]*URI, error) {
	args := m.MethodCalled("Discover", u)

	if tmp := args.Get(0); tmp != nil {
		return tmp.([]*URI), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestRegisterDiscovery(t *testing.T) {
	mech := &mockDiscovery{}
	defer patcher.SetVar(&discMechs, map[string]Discovery{}).Install().Restore()

	RegisterDiscovery("test", mech)

	assert.Equal(t, map[string]Discovery{
		"test": mech,
	}, discMechs)
}

func TestLookupDiscovery(t *testing.T) {
	mech := &mockDiscovery{}
	defer patcher.SetVar(&discMechs, map[string]Discovery{
		"test": mech,
	}).Install().Restore()

	result := LookupDiscovery("test")

	assert.Same(t, mech, result)
}

type mockMechanism struct {
	mock.Mock
}

func (m *mockMechanism) Dial(config Config, u *URI) (*Conduit, error) {
	args := m.MethodCalled("Dial", config, u)

	if tmp := args.Get(0); tmp != nil {
		return tmp.(*Conduit), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *mockMechanism) Listen(config Config, u *URI) (Listener, error) {
	args := m.MethodCalled("Listen", config, u)

	if tmp := args.Get(0); tmp != nil {
		return tmp.(Listener), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestRegisterSecurity(t *testing.T) {
	mech := &mockMechanism{}
	defer patcher.SetVar(&secMechs, map[string]Mechanism{}).Install().Restore()

	RegisterSecurity("test", mech)

	assert.Equal(t, map[string]Mechanism{
		"test": mech,
	}, secMechs)
}

func TestLookupSecurity(t *testing.T) {
	mech := &mockMechanism{}
	defer patcher.SetVar(&secMechs, map[string]Mechanism{
		"test": mech,
	}).Install().Restore()

	result := LookupSecurity("test")

	assert.Same(t, mech, result)
}

func TestRegisterTransport(t *testing.T) {
	mech := &mockMechanism{}
	defer patcher.SetVar(&transMechs, map[string]Mechanism{}).Install().Restore()

	RegisterTransport("test", mech)

	assert.Equal(t, map[string]Mechanism{
		"test": mech,
	}, transMechs)
}

func TestLookupTransport(t *testing.T) {
	mech := &mockMechanism{}
	defer patcher.SetVar(&transMechs, map[string]Mechanism{
		"test": mech,
	}).Install().Restore()

	result := LookupTransport("test")

	assert.Same(t, mech, result)
}
