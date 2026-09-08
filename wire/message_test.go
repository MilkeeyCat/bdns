package wire_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/message"
	"github.com/MilkeeyCat/bdns/record"
	"github.com/MilkeeyCat/bdns/wire"
)

func TestMessage(t *testing.T) {
	tests := []struct {
		input   []byte
		message *message.Message
		err     error
	}{
		{
			// dig @127.0.0.1 -p 8080 deez.local. +noedns
			input: []byte{0x0d, 0x3d, 0x01, 0x20, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x64, 0x65, 0x65, 0x7a, 0x05, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x00, 0x00, 0x01, 0x00, 0x01},
			message: &message.Message{
				Header: message.Header{
					ID:     0x0d3d,
					QR:     false,
					Opcode: message.QueryOpcodeStandard,
					AA:     false,
					TC:     false,
					RD:     true,
					RA:     false,
					RCode:  message.ResponseCodeNone,
				},
				Question: []message.Question{
					{
						Name:  domain.Domain{{'d', 'e', 'e', 'z'}, {'l', 'o', 'c', 'a', 'l'}, {}},
						Type:  message.QueryType(record.TypeA),
						Class: message.QueryClass(record.ClassIN),
					},
				},
				Answer:     []record.Record{},
				Authority:  []record.Record{},
				Additional: []record.Record{},
			},
			err: nil,
		},
		{
			input: []byte{0xb6, 0x49, 0x81, 0xa0, 0x00, 0x01, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x01, 0x00, 0x01, 0xc0, 0x0c, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x2c, 0x00, 0x04, 0xac, 0x42, 0x93, 0xf3, 0xc0, 0x0c, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x2c, 0x00, 0x04, 0x68, 0x14, 0x17, 0x9a},
			message: &message.Message{
				Header: message.Header{
					ID:     0xb649,
					QR:     true,
					Opcode: message.QueryOpcodeStandard,
					AA:     false,
					TC:     false,
					RD:     true,
					RA:     true,
					RCode:  message.ResponseCodeNone,
				},
				Question: []message.Question{
					{
						Name:  domain.Domain{{'e', 'x', 'a', 'm', 'p', 'l', 'e'}, {'c', 'o', 'm'}, {}},
						Type:  message.QueryType(record.TypeA),
						Class: message.QueryClass(record.ClassIN),
					},
				},
				Answer: []record.Record{
					{
						Name:  domain.Domain{{'e', 'x', 'a', 'm', 'p', 'l', 'e'}, {'c', 'o', 'm'}, {}},
						Type:  record.TypeA,
						Class: record.ClassIN,
						TTL:   300,
						Data:  []byte{172, 66, 147, 243},
					},
					{
						Name:  domain.Domain{{'e', 'x', 'a', 'm', 'p', 'l', 'e'}, {'c', 'o', 'm'}, {}},
						Type:  record.TypeA,
						Class: record.ClassIN,
						TTL:   300,
						Data:  []byte{104, 20, 23, 154},
					},
				},
				Authority:  []record.Record{},
				Additional: []record.Record{},
			},
			err: nil,
		},
	}

	for _, tc := range tests {
		message, err := wire.ParseMessage(tc.input)

		if tc.message != nil {
			assert.Equal(t, *tc.message, message)
		}

		assert.Equal(t, tc.err, err)
	}
}
