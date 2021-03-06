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
	"syscall"
)

// Patch points for isolating functions during testing.
var (
	lookupIP            func(string) ([]net.IP, error)                                          = net.LookupIP
	lookupPort          func(string, string) (int, error)                                       = net.LookupPort
	lookupSecurity      func(string) Mechanism                                                  = LookupSecurity
	lookupTransport     func(string) Mechanism                                                  = LookupTransport
	mkDialerPatch       func(opts []DialerOption, filt dialerFilter) (iDialer, error)           = mkDialer
	mkListenConfigPatch func(opts []ListenerOption, filt listenerFilter) (iListenConfig, error) = mkListenConfig
	setsockoptInt       func(fd, level, opt, value int) error                                   = syscall.SetsockoptInt
	resolveTCPAddr      func(network, address string) (*net.TCPAddr, error)                     = net.ResolveTCPAddr
)
