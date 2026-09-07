package wire_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MilkeeyCat/bdns/message"
	"github.com/MilkeeyCat/bdns/wire"
)

func TestParseHeader(t *testing.T) {
	tests := []struct {
		input  [wire.HeaderSize]byte
		header *wire.Header
		err    error
	}{
		{
			// dig @127.0.0.1 -p 8080 .
			input: [wire.HeaderSize]byte{0xce, 0x5d, 0x01, 0x20, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			header: &wire.Header{
				ID:      0xce5d,
				QR:      false,
				Opcode:  message.QueryOpcodeStandard,
				AA:      false,
				TC:      false,
				RD:      true,
				RA:      false,
				RCode:   0,
				QDCount: 1,
				ANCount: 0,
				NSCount: 0,
				ARCount: 1,
			},
			err: nil,
		},
		{
			// dig @127.0.0.1 -p 8080 +aaflag +tcflag +recurse +raflag .
			input: [wire.HeaderSize]byte{0x6a, 0x4c, 0x07, 0xa0, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			header: &wire.Header{
				ID:      0x6a4c,
				QR:      false,
				Opcode:  message.QueryOpcodeStandard,
				AA:      true,
				TC:      true,
				RD:      true,
				RA:      true,
				RCode:   0,
				QDCount: 1,
				ANCount: 0,
				NSCount: 0,
				ARCount: 1,
			},
			err: nil,
		},
		{
			input:  [wire.HeaderSize]byte{0xce, 0x5d, 0x79, 0x20, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10},
			header: nil,
			err:    wire.ErrInvalidMessage,
		},
		{
			input:  [wire.HeaderSize]byte{0xce, 0x5d, 0x01, 0x2f, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			header: nil,
			err:    wire.ErrInvalidMessage,
		},
	}

	for _, tc := range tests {
		header, err := wire.ParseHeader(tc.input)

		if tc.header != nil {
			assert.Equal(t, *tc.header, header)
		}

		assert.Equal(t, tc.err, err)
	}
}
