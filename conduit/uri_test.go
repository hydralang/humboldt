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
	"errors"
	"net"
	"net/url"
	"testing"

	"github.com/klmitch/patcher"
	"github.com/stretchr/testify/assert"
)

func TestParseBase(t *testing.T) {
	result, err := Parse("tcp://127.0.0.1:1234")

	assert.NoError(t, err)
	assert.Equal(t, "tcp", result.Scheme)
	assert.Equal(t, "127.0.0.1:1234", result.Host)
	assert.Equal(t, "tcp", result.Transport)
	assert.Equal(t, "", result.Security)
	assert.Equal(t, "", result.Discovery)
}

func TestParsePortOnly(t *testing.T) {
	result, err := Parse("tcp://:1234")

	assert.NoError(t, err)
	assert.Equal(t, "tcp", result.Scheme)
	assert.Equal(t, ":1234", result.Host)
	assert.Equal(t, "tcp", result.Transport)
	assert.Equal(t, "", result.Security)
	assert.Equal(t, "", result.Discovery)
}

func TestParseDiscovery(t *testing.T) {
	result, err := Parse("tcp.srv://127.0.0.1:1234")

	assert.NoError(t, err)
	assert.Equal(t, "tcp.srv", result.Scheme)
	assert.Equal(t, "127.0.0.1:1234", result.Host)
	assert.Equal(t, "tcp", result.Transport)
	assert.Equal(t, "", result.Security)
	assert.Equal(t, "srv", result.Discovery)
}

func TestParseSecurity(t *testing.T) {
	result, err := Parse("tcp+tls://127.0.0.1:1234")

	assert.NoError(t, err)
	assert.Equal(t, "tcp+tls", result.Scheme)
	assert.Equal(t, "127.0.0.1:1234", result.Host)
	assert.Equal(t, "tcp", result.Transport)
	assert.Equal(t, "tls", result.Security)
	assert.Equal(t, "", result.Discovery)
}

func TestParsePathological(t *testing.T) {
	result, err := Parse("tcp+s1+s2.d1.d2://127.0.0.1:1234")

	assert.NoError(t, err)
	assert.Equal(t, "tcp+s1+s2.d1.d2", result.Scheme)
	assert.Equal(t, "127.0.0.1:1234", result.Host)
	assert.Equal(t, "tcp", result.Transport)
	assert.Equal(t, "s1+s2.d1", result.Security)
	assert.Equal(t, "d2", result.Discovery)
}

func TestParseError(t *testing.T) {
	result, err := Parse("://127.0.0.1:1234")

	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestURIIsCanonicalBase(t *testing.T) {
	obj := &URI{}

	result := obj.IsCanonical()

	assert.True(t, result)
}

func TestURIIsCanonicalWithDiscovery(t *testing.T) {
	obj := &URI{
		Discovery: "srv",
	}

	result := obj.IsCanonical()

	assert.False(t, result)
}

func TestURIIsCanonicalIPOnly(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1",
		},
	}

	result := obj.IsCanonical()

	assert.False(t, result)
}

func TestURIIsCanonicalTextPort(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:humboldt",
		},
	}

	result := obj.IsCanonical()

	assert.False(t, result)
}

func TestURIIsCanonicalNumericPort(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
	}

	result := obj.IsCanonical()

	assert.True(t, result)
}

func TestURIIsCanonicalHostName(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "localhost:1234",
		},
	}

	result := obj.IsCanonical()

	assert.False(t, result)
}

func TestURICanonicalizeBase(t *testing.T) {
	obj := &URI{}

	result, err := obj.Canonicalize()

	assert.NoError(t, err)
	assert.Equal(t, []*URI{
		obj,
	}, result)
}

func TestURICanonicalizeDiscoveryMissing(t *testing.T) {
	obj := &URI{
		Discovery: "missing",
	}
	defer patcher.SetVar(&discMechs, map[string]Discovery{}).Install().Restore()

	result, err := obj.Canonicalize()

	assert.True(t, errors.Is(err, ErrUnknownDiscovery))
	assert.Nil(t, result)
}

func TestURICanonicalizeDiscovery(t *testing.T) {
	disc := &mockDiscovery{}
	obj := &URI{
		Discovery: "disc",
	}
	disc.On("Discover", obj).Return([]*URI{obj}, assert.AnError)
	defer patcher.SetVar(&discMechs, map[string]Discovery{
		"disc": disc,
	}).Install().Restore()

	result, err := obj.Canonicalize()

	assert.Same(t, assert.AnError, err)
	assert.Equal(t, []*URI{obj}, result)
	disc.AssertExpectations(t)
}

func TestURICanonicalizeCanonical(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
	}

	result, err := obj.Canonicalize()

	assert.NoError(t, err)
	assert.Equal(t, []*URI{
		obj,
	}, result)
}

func TestURICanonicalizeNoPort(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1",
		},
	}

	result, err := obj.Canonicalize()

	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestURICanonicalizeHostname(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "localhost:1234",
		},
	}
	defer patcher.SetVar(&lookupIP, func(host string) ([]net.IP, error) {
		assert.Equal(t, "localhost", host)
		return []net.IP{
			net.IPv4(127, 0, 0, 1),
			net.IPv6loopback,
		}, nil
	}).Install().Restore()

	result, err := obj.Canonicalize()

	assert.NoError(t, err)
	assert.Equal(t, []*URI{
		{
			URL: url.URL{
				Host: "127.0.0.1:1234",
			},
		},
		{
			URL: url.URL{
				Host: "[::1]:1234",
			},
		},
	}, result)
}

func TestURICanonicalizeHostnameLookupFails(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "localhost:1234",
		},
	}
	defer patcher.SetVar(&lookupIP, func(host string) ([]net.IP, error) {
		assert.Equal(t, "localhost", host)
		return nil, assert.AnError
	}).Install().Restore()

	result, err := obj.Canonicalize()

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
}

func TestURICanonicalizePort(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:humboldt",
		},
		Transport: "tcp",
	}
	defer patcher.SetVar(&lookupPort, func(network, port string) (int, error) {
		assert.Equal(t, "tcp", network)
		assert.Equal(t, "humboldt", port)
		return 1234, nil
	}).Install().Restore()

	result, err := obj.Canonicalize()

	assert.NoError(t, err)
	assert.Equal(t, []*URI{
		{
			URL: url.URL{
				Host: "127.0.0.1:1234",
			},
			Transport: "tcp",
		},
	}, result)
}

func TestURICanonicalizePortLookupFails(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:humboldt",
		},
		Transport: "tcp",
	}
	defer patcher.SetVar(&lookupPort, func(network, port string) (int, error) {
		assert.Equal(t, "tcp", network)
		assert.Equal(t, "humboldt", port)
		return 0, assert.AnError
	}).Install().Restore()

	result, err := obj.Canonicalize()

	assert.Same(t, assert.AnError, err)
	assert.Nil(t, result)
}
