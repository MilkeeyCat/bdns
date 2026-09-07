package wire_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MilkeeyCat/bdns/wire"
)

func TestDomain(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		offset uint
		domain *string
		size   *uint8
		err    error
	}{
		{
			name:   "not compressed",
			input:  []byte{0x01, 'f', 0x03, 'i', 's', 'i', 0x00},
			offset: 0,
			domain: new("f.isi."),
			size:   new(uint8(7)),
			err:    nil,
		},
		{
			name: "compressed",
			input: []byte{
				0x01, 'm', 0x03, 'i', 's', 'i', 0x04, 'a', 'r', 'p', 'a', 0x00,
				0x03, 'f', 'o', 'o', 0xc0, 0x00,
			},
			offset: 12,
			domain: new("foo.m.isi.arpa."),
			size:   new(uint8(6)),
			err:    nil,
		},
		{
			name:   "compressed with cycle",
			input:  []byte{0x03, 'f', 'o', 'o', 0xc0, 0x00},
			offset: 0,
			domain: nil,
			size:   nil,
			err:    wire.ErrInvalidMessage,
		},
		{
			name:   "compressed with offset out of bounds",
			input:  []byte{0x03, 'f', 'o', 'o', 0xc0, 0xff},
			offset: 0,
			domain: nil,
			size:   nil,
			err:    wire.ErrInvalidMessage,
		},
		{
			name:   "invalid high bits of size byte",
			input:  []byte{0xc3, 'f', 'o', 'o', 0xc0, 0xff},
			offset: 0,
			domain: nil,
			size:   nil,
			err:    wire.ErrInvalidMessage,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			domain, size, err := wire.ParseDomain(tc.input, tc.offset)

			if tc.domain != nil {
				assert.Equal(t, *tc.domain, domain.String())
				assert.Equal(t, *tc.size, size)
			}

			assert.Equal(t, tc.err, err)
		})
	}
}
