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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConduitSend(t *testing.T) {
	link := &mockLink{}
	link.On("Send", []byte("msg")).Return(assert.AnError)
	obj := Conduit{
		link: link,
	}

	err := obj.Send([]byte("msg"))

	assert.Same(t, assert.AnError, err)
	link.AssertExpectations(t)
}

func TestConduitClose(t *testing.T) {
	link := &mockLink{}
	link.On("Close")
	obj := Conduit{
		link: link,
	}

	obj.Close()

	link.AssertExpectations(t)
}
