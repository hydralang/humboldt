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
	"context"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hydralang/humboldt/conduit"
)

type Server struct {
	sync.Mutex
	Wg        *sync.WaitGroup
	Listener  conduit.Listener
	URI       string
	Data      map[string][][]byte
	Errors    map[string]error
	AcceptErr error
}

func NewServer(cfg conduit.Config, uri string) (*Server, error) {
	l, err := conduit.Listen(context.Background(), cfg, uri)
	if err != nil {
		return nil, err
	}

	return &Server{
		Wg:       &sync.WaitGroup{},
		Listener: l,
		URI:      l.Addr().String(),
		Data:     map[string][][]byte{},
		Errors:   map[string]error{},
	}, nil
}

func (s *Server) Echo(c *conduit.Conduit) {
	defer s.Wg.Done()
	uri := c.RemoteURI.String()
	buf := make([]byte, 1024)
	for {
		inLen, err := c.Link.Read(buf)
		if err != nil {
			if err != io.EOF {
				s.Lock()
				s.Errors[uri] = err
				s.Unlock()
			}
			break
		}
		tmp := make([]byte, inLen)
		copy(tmp, buf)
		s.Lock()
		s.Data[uri] = append(s.Data[uri], tmp)
		s.Unlock()
		_, err = c.Link.Write(buf[:inLen])
		if err != nil {
			s.Lock()
			s.Errors[uri] = err
			s.Unlock()
			break
		}
	}
}

func (s *Server) Accept() {
	defer s.Wg.Done()
	for {
		c, err := s.Listener.Accept()
		if err != nil {
			s.Lock()
			s.AcceptErr = err
			s.Unlock()
			break
		}
		s.Wg.Add(1)
		go s.Echo(c)
	}
}

func (s *Server) Start() {
	s.Wg.Add(1)
	go s.Accept()
}

func (s *Server) Close() {
	s.Lock()
	if s.Listener != nil {
		s.Listener.Close()
		s.Listener = nil
	}
	s.Unlock()
	s.Wg.Wait()
}

type Client struct {
	Wg      *sync.WaitGroup
	Conduit *conduit.Conduit
	URI     string
	Out     [][]byte
	In      [][]byte
	Error   error
}

func NewClient(wg *sync.WaitGroup, cfg conduit.Config, uri string, data [][]byte) (*Client, error) {
	c, err := conduit.Dial(context.Background(), cfg, uri)
	if err != nil {
		return nil, err
	}

	return &Client{
		Wg:      wg,
		Conduit: c,
		URI:     c.LocalURI.String(),
		Out:     data,
	}, nil
}

func (c *Client) Send() {
	defer c.Wg.Done()
	defer c.Conduit.Link.Close()
	inBuf := make([]byte, 1024)
	for _, buf := range c.Out {
		_, err := c.Conduit.Link.Write(buf)
		if err != nil {
			c.Error = err
			return
		}

		inLen, err := c.Conduit.Link.Read(inBuf)
		if err != nil {
			c.Error = err
			return
		}
		tmp := make([]byte, inLen)
		copy(tmp, inBuf)
		c.In = append(c.In, tmp)
	}
}

func (c *Client) Start() {
	c.Wg.Add(1)
	go c.Send()
}

type Config struct {
	Transport map[string]interface{}
	Security  map[string]interface{}
}

func (c *Config) ForTransport(name string) interface{} {
	return c.Transport[name]
}

func (c *Config) ForSecurity(name string) interface{} {
	return c.Security[name]
}

type Scenario struct {
	URI  string   // Listen URI of the server
	Cli1 [][]byte // Send data for client 1
	Cli2 [][]byte // Send data for client 2
	Cfg  *Config  // Configuration
}

func (s *Scenario) Execute(t *testing.T) {
	// Construct and start the server
	server, err := NewServer(s.Cfg, s.URI)
	require.NoError(t, err)
	server.Start()
	defer server.Close()

	// Construct the clients
	wg := &sync.WaitGroup{}
	cli1, err := NewClient(wg, s.Cfg, server.URI, s.Cli1)
	require.NoError(t, err)
	cli2, err := NewClient(wg, s.Cfg, server.URI, s.Cli2)
	require.NoError(t, err)

	// Start them
	cli1.Start()
	cli2.Start()

	// Wait for clients to finish
	wg.Wait()

	// Shut down the server
	server.Close()

	// Check the clients
	assert.Equal(t, server.URI, cli1.Conduit.RemoteURI.String())
	assert.Equal(t, s.Cli1, cli1.In)
	assert.NoError(t, cli1.Error)
	assert.Equal(t, server.URI, cli2.Conduit.RemoteURI.String())
	assert.Equal(t, s.Cli2, cli2.In)
	assert.NoError(t, cli2.Error)

	// Check the server
	cli1URI := cli1.Conduit.LocalURI.String()
	require.Contains(t, server.Data, cli1URI)
	assert.Equal(t, s.Cli1, server.Data[cli1URI])
	cli2URI := cli2.Conduit.LocalURI.String()
	require.Contains(t, server.Data, cli2URI)
	assert.Equal(t, s.Cli2, server.Data[cli2URI])
}
