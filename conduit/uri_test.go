// Copyright (c) 2020, 2021 Kevin L. Mitchell
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

	assert.Error(t, err)
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

	assert.ErrorIs(t, err, ErrUnknownDiscovery)
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

	assert.Error(t, err)
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

func TestURIDialBase(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	c := &Conduit{}
	mech.On("Dial", cfg, obj).Return(c, nil)
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return nil
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return mech
		}),
	).Install().Restore()

	result, err := obj.Dial(cfg)

	assert.NoError(t, err)
	assert.Same(t, c, result)
	mech.AssertExpectations(t)
	assert.False(t, securityCalled)
	assert.True(t, transportCalled)
}

func TestURIDialWithSecurity(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
		Security:  "tls",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	c := &Conduit{}
	mech.On("Dial", cfg, obj).Return(c, nil)
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return mech
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return nil
		}),
	).Install().Restore()

	result, err := obj.Dial(cfg)

	assert.NoError(t, err)
	assert.Same(t, c, result)
	mech.AssertExpectations(t)
	assert.True(t, securityCalled)
	assert.False(t, transportCalled)
}

func TestURIDialNotCanonical(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
		Discovery: "srv",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return nil
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return mech
		}),
	).Install().Restore()

	result, err := obj.Dial(cfg)

	assert.ErrorIs(t, err, ErrNotCanonical)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
	assert.False(t, securityCalled)
	assert.False(t, transportCalled)
}

func TestURIDialBlankTransport(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return nil
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return mech
		}),
	).Install().Restore()

	result, err := obj.Dial(cfg)

	assert.ErrorIs(t, err, ErrUnknownTransport)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
	assert.False(t, securityCalled)
	assert.False(t, transportCalled)
}

func TestURIDialUnknownSecurity(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
		Security:  "tls",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return nil
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return mech
		}),
	).Install().Restore()

	result, err := obj.Dial(cfg)

	assert.ErrorIs(t, err, ErrUnknownSecurity)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
	assert.True(t, securityCalled)
	assert.False(t, transportCalled)
}

func TestURIDialUnknownTransport(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return mech
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return nil
		}),
	).Install().Restore()

	result, err := obj.Dial(cfg)

	assert.ErrorIs(t, err, ErrUnknownTransport)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
	assert.False(t, securityCalled)
	assert.True(t, transportCalled)
}

func TestURIListenBase(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	l := &mockListener{}
	mech.On("Listen", cfg, obj).Return(l, nil)
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return nil
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return mech
		}),
	).Install().Restore()

	result, err := obj.Listen(cfg)

	assert.NoError(t, err)
	assert.Same(t, l, result)
	mech.AssertExpectations(t)
	assert.False(t, securityCalled)
	assert.True(t, transportCalled)
}

func TestURIListenWithSecurity(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
		Security:  "tls",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	l := &mockListener{}
	mech.On("Listen", cfg, obj).Return(l, nil)
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return mech
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return nil
		}),
	).Install().Restore()

	result, err := obj.Listen(cfg)

	assert.NoError(t, err)
	assert.Same(t, l, result)
	mech.AssertExpectations(t)
	assert.True(t, securityCalled)
	assert.False(t, transportCalled)
}

func TestURIListenNotCanonical(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
		Discovery: "srv",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return nil
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return mech
		}),
	).Install().Restore()

	result, err := obj.Listen(cfg)

	assert.ErrorIs(t, err, ErrNotCanonical)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
	assert.False(t, securityCalled)
	assert.False(t, transportCalled)
}

func TestURIListenBlankTransport(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return nil
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return mech
		}),
	).Install().Restore()

	result, err := obj.Listen(cfg)

	assert.ErrorIs(t, err, ErrUnknownTransport)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
	assert.False(t, securityCalled)
	assert.False(t, transportCalled)
}

func TestURIListenUnknownSecurity(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
		Security:  "tls",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return nil
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return mech
		}),
	).Install().Restore()

	result, err := obj.Listen(cfg)

	assert.ErrorIs(t, err, ErrUnknownSecurity)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
	assert.True(t, securityCalled)
	assert.False(t, transportCalled)
}

func TestURIListenUnknownTransport(t *testing.T) {
	obj := &URI{
		URL: url.URL{
			Host: "127.0.0.1:1234",
		},
		Transport: "tcp",
	}
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	securityCalled := false
	transportCalled := false
	defer patcher.NewPatchMaster(
		patcher.SetVar(&lookupSecurity, func(name string) Mechanism {
			assert.Equal(t, "tls", name)
			securityCalled = true
			return mech
		}),
		patcher.SetVar(&lookupTransport, func(name string) Mechanism {
			assert.Equal(t, "tcp", name)
			transportCalled = true
			return nil
		}),
	).Install().Restore()

	result, err := obj.Listen(cfg)

	assert.ErrorIs(t, err, ErrUnknownTransport)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
	assert.False(t, securityCalled)
	assert.True(t, transportCalled)
}

func TestDialBase(t *testing.T) {
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	c := &Conduit{}
	mech.On("Dial", cfg, &URI{
		URL: url.URL{
			Scheme: "tcp",
			Host:   "127.0.0.1:1234",
		},
		Transport: "tcp",
	}).Return(c, nil)
	defer patcher.SetVar(&lookupTransport, func(name string) Mechanism {
		return mech
	}).Install().Restore()

	result, err := Dial(cfg, "tcp://127.0.0.1:1234")

	assert.NoError(t, err)
	assert.Same(t, c, result)
	mech.AssertExpectations(t)
}

func TestDialParseError(t *testing.T) {
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	defer patcher.SetVar(&lookupTransport, func(name string) Mechanism {
		return mech
	}).Install().Restore()

	result, err := Dial(cfg, "://127.0.0.1:1234")

	assert.Error(t, err)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
}

func TestListenBase(t *testing.T) {
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	l := &mockListener{}
	mech.On("Listen", cfg, &URI{
		URL: url.URL{
			Scheme: "tcp",
			Host:   "127.0.0.1:1234",
		},
		Transport: "tcp",
	}).Return(l, nil)
	defer patcher.SetVar(&lookupTransport, func(name string) Mechanism {
		return mech
	}).Install().Restore()

	result, err := Listen(cfg, "tcp://127.0.0.1:1234")

	assert.NoError(t, err)
	assert.Same(t, l, result)
	mech.AssertExpectations(t)
}

func TestListenParseError(t *testing.T) {
	mech := &mockMechanism{}
	cfg := &mock.Mock{}
	defer patcher.SetVar(&lookupTransport, func(name string) Mechanism {
		return mech
	}).Install().Restore()

	result, err := Listen(cfg, "://127.0.0.1:1234")

	assert.Error(t, err)
	assert.Nil(t, result)
	mech.AssertExpectations(t)
}
