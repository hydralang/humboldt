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

import "github.com/stretchr/testify/mock"

type mockListener struct {
	mock.Mock
}

func (m *mockListener) Accept() (*Conduit, error) {
	args := m.MethodCalled("Accept")

	if tmp := args.Get(0); tmp != nil {
		return tmp.(*Conduit), args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *mockListener) Close() error {
	args := m.MethodCalled("Close")

	return args.Error(0)
}

func (m *mockListener) Addr() *URI {
	args := m.MethodCalled("Addr")

	if tmp := args.Get(0); tmp != nil {
		return tmp.(*URI)
	}

	return nil
}
