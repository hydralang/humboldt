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

package proto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderFromBytesBase(t *testing.T) {
	obj := &Header{}
	data := []byte{
		0x08,
		0x17,
		0x01, 0xff,
	}

	result, err := obj.FromBytes(data)

	assert.NoError(t, err)
	assert.Equal(t, HeaderSize, result)
	assert.Equal(t, &Header{
		Major:    0x00,
		Reply:    true,
		Error:    false,
		Protocol: 0x17,
		Length:   0x01ff,
	}, obj)
}

func TestHeaderFromBytesShort(t *testing.T) {
	obj := &Header{}
	data := []byte{
		0x08,
		0x17,
		0x01,
	}

	result, err := obj.FromBytes(data)

	assert.ErrorIs(t, err, ErrShortInput)
	assert.Equal(t, 0, result)
	assert.Equal(t, &Header{}, obj)
}

func TestHeaderFromBytesHighMajor(t *testing.T) {
	obj := &Header{}
	data := []byte{
		(MaxMajor+1)<<MajorShift | 0x08,
		0x17,
		0x01, 0xff,
	}

	result, err := obj.FromBytes(data)

	assert.ErrorIs(t, err, ErrMaxVersion)
	assert.Equal(t, 0, result)
	assert.Equal(t, &Header{}, obj)
}

func TestHeaderToBytesBase(t *testing.T) {
	obj := &Header{
		Major:    0x00,
		Reply:    true,
		Error:    true,
		Protocol: 0x17,
		Length:   0x01ff,
	}
	buf := make([]byte, 6)

	result, err := obj.ToBytes(buf)

	assert.NoError(t, err)
	assert.Equal(t, HeaderSize, result)
	assert.Equal(t, []byte{
		0x0c,
		0x17,
		0x01, 0xff,
		0x00, 0x00,
	}, buf)
}

func TestHeaderToBytesSmall(t *testing.T) {
	obj := &Header{
		Major:    0x00,
		Reply:    true,
		Error:    true,
		Protocol: 0x17,
		Length:   0x01ff,
	}
	buf := make([]byte, 3, 6)

	result, err := obj.ToBytes(buf)

	assert.ErrorIs(t, err, ErrShortOutput)
	assert.Equal(t, 0, result)
	assert.Equal(t, []byte{
		0x00,
		0x00,
		0x00,
	}, buf)
}

func TestHeaderToBytesHighMajor(t *testing.T) {
	obj := &Header{
		Major:    MaxMajor + 1,
		Reply:    true,
		Error:    true,
		Protocol: 0x17,
		Length:   0x01ff,
	}
	buf := make([]byte, 6)

	result, err := obj.ToBytes(buf)

	assert.ErrorIs(t, err, ErrMaxVersion)
	assert.Equal(t, 0, result)
	assert.Equal(t, []byte{
		0x00,
		0x00,
		0x00, 0x00,
		0x00, 0x00,
	}, buf)
}

func TestExtHeaderFromBytesBase(t *testing.T) {
	obj := &ExtHeader{}
	data := []byte{
		0xa0,
		0x17,
		0x01, 0xff,
	}

	result, err := obj.FromBytes(data)

	assert.NoError(t, err)
	assert.Equal(t, ExtHeaderSize, result)
	assert.Equal(t, &ExtHeader{
		Ignore:   true,
		Close:    false,
		HopByHop: true,
		Protocol: 0x17,
		Length:   0x01ff,
	}, obj)
}

func TestExtHeaderFromBytesShort(t *testing.T) {
	obj := &ExtHeader{}
	data := []byte{
		0xa0,
		0x17,
		0x01,
	}

	result, err := obj.FromBytes(data)

	assert.ErrorIs(t, err, ErrShortInput)
	assert.Equal(t, 0, result)
	assert.Equal(t, &ExtHeader{}, obj)
}

func TestExtHeaderToBytesBase(t *testing.T) {
	obj := &ExtHeader{
		Ignore:   true,
		Close:    true,
		HopByHop: true,
		Protocol: 0x17,
		Length:   0x01ff,
	}
	buf := make([]byte, 6)

	result, err := obj.ToBytes(buf)

	assert.NoError(t, err)
	assert.Equal(t, ExtHeaderSize, result)
	assert.Equal(t, []byte{
		0xe0,
		0x17,
		0x01, 0xff,
		0x00, 0x00,
	}, buf)
}

func TestExtHeaderToBytesSmall(t *testing.T) {
	obj := &ExtHeader{
		Ignore:   true,
		Close:    true,
		HopByHop: true,
		Protocol: 0x17,
		Length:   0x01ff,
	}
	buf := make([]byte, 3, 6)

	result, err := obj.ToBytes(buf)

	assert.ErrorIs(t, err, ErrShortOutput)
	assert.Equal(t, 0, result)
	assert.Equal(t, []byte{
		0x00,
		0x00,
		0x00,
	}, buf)
}
