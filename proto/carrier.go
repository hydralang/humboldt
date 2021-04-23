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

import "fmt"

// Constants used in the binary encoding of Header.
const (
	HeaderSize int   = 4
	MaxMajor   uint8 = 0
	MajorMask  uint8 = 0xf0
	MajorShift int   = 4
	ReplyBit   uint8 = 0x08
	ErrorBit   uint8 = 0x04
)

// Header describes the header of a Humboldt PDU.
type Header struct {
	Major    uint8  // Major protocol version
	Reply    bool   // Reply flag
	Error    bool   // Error flag
	Protocol uint8  // Protocol number
	Length   uint16 // Packet length
}

// FromBytes is a method of Header that fills in the information from
// a sequence of 4 bytes.
func (h *Header) FromBytes(data []byte) (int, error) {
	// Make sure we have enough data
	if len(data) < HeaderSize {
		return 0, ErrShortInput
	}

	// Check the version
	vers := (data[0] & MajorMask) >> MajorShift
	if vers > MaxMajor {
		return 0, fmt.Errorf("%d: %w", vers, ErrMaxVersion)
	}

	// Fill in the header
	h.Major = vers
	h.Reply = (data[0] & ReplyBit) != 0
	h.Error = (data[0] & ErrorBit) != 0
	h.Protocol = data[1]
	h.Length = (uint16(data[2]) << 8) | uint16(data[3])

	return HeaderSize, nil
}

// ToBytes is a method of Header that encodes the header into a
// sequence of 4 bytes.  The byte slice to fill in must be passed in.
func (h *Header) ToBytes(data []byte) (int, error) {
	// Make sure the version makes sense
	if h.Major > MaxMajor {
		return 0, fmt.Errorf("%d: %w", h.Major, ErrMaxVersion)
	}

	// Make sure we have enough space
	if len(data) < HeaderSize {
		return 0, ErrShortOutput
	}

	// Fill in the data
	data[0] = (h.Major << MajorShift) & MajorMask
	if h.Reply {
		data[0] |= ReplyBit
	}
	if h.Error {
		data[0] |= ErrorBit
	}
	data[1] = h.Protocol
	data[2] = uint8((h.Length & 0xff00) >> 8)
	data[3] = uint8(h.Length & 0x00ff)

	return HeaderSize, nil
}

// Constants used in the binary encoding of ExtensionHeader
const (
	ExtHeaderSize int   = 4
	IgnoreBit     uint8 = 0x80
	CloseBit      uint8 = 0x40
	HopBit        uint8 = 0x20
)

// ExtHeader describes the header of a Humboldt protocol extension.
type ExtHeader struct {
	Ignore   bool   // Ignore flag
	Close    bool   // Close flag
	HopByHop bool   // Hop-by-hop flag
	Protocol uint8  // Next protocol number
	Length   uint16 // Extension length
}

// FromBytes is a method of ExtHeader that fills in the information
// from a sequence of 4 bytes.
func (h *ExtHeader) FromBytes(data []byte) (int, error) {
	// Make sure we have enough data
	if len(data) < ExtHeaderSize {
		return 0, ErrShortInput
	}

	// Fill in the header
	h.Ignore = (data[0] & IgnoreBit) != 0
	h.Close = (data[0] & CloseBit) != 0
	h.HopByHop = (data[0] & HopBit) != 0
	h.Protocol = data[1]
	h.Length = (uint16(data[2]) << 8) | uint16(data[3])

	return ExtHeaderSize, nil
}

// ToBytes is a method of ExtHeader that encodes the header into a
// sequence of 4 bytes.  The byte slice to fill in must be passed in.
func (h *ExtHeader) ToBytes(data []byte) (int, error) {
	// Make sure we have enough space
	if len(data) < ExtHeaderSize {
		return 0, ErrShortOutput
	}

	// Fill in the data
	if h.Ignore {
		data[0] |= IgnoreBit
	}
	if h.Close {
		data[0] |= CloseBit
	}
	if h.HopByHop {
		data[0] |= HopBit
	}
	data[1] = h.Protocol
	data[2] = uint8((h.Length & 0xff00) >> 8)
	data[3] = uint8(h.Length & 0x00ff)

	return ExtHeaderSize, nil
}
