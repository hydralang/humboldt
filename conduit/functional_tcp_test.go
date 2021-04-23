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

package conduit_test

import (
	"testing"
)

func TestTCP(t *testing.T) {
	s := &Scenario{
		URI:  "tcp://127.0.0.1:0",
		Cli1: [][]byte{[]byte("test"), []byte("one\n"), []byte("two\r\n")},
		Cli2: [][]byte{[]byte("test2"), []byte("three\r"), []byte("four")},
	}

	s.Execute(t)
}