package wire_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/message"
	"github.com/MilkeeyCat/bdns/record"
	"github.com/MilkeeyCat/bdns/wire"
)

func TestParser(t *testing.T) {
	t.Run("domain", testParseDomain)
	t.Run("question", testParseQuestion)
	t.Run("resource record", testParseResourceRecord)
	t.Run("message", testParseMessage)
}

func testParseDomain(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		offset uint
		domain *string
		size   *uint
		err    error
	}{
		{
			name:   "not compressed",
			input:  []byte{0x01, 'f', 0x03, 'i', 's', 'i', 0x00},
			offset: 0,
			domain: new("f.isi."),
			size:   new(uint(7)),
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
			size:   new(uint(6)),
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
			parser := wire.NewParser(tc.input, wire.WithOffset(tc.offset))
			domain, err := parser.ParseDomain()

			assert.Equal(t, tc.err, err)

			if tc.domain != nil {
				assert.Equal(t, *tc.domain, domain.String())
				assert.Equal(t, *tc.size, parser.Offset()-tc.offset)
			}
		})
	}
}

func testParseQuestion(t *testing.T) {
	tests := []struct {
		input    []byte
		offset   uint
		question *message.Question
		size     *uint
		err      error
	}{
		{
			// dig @127.0.0.1 -p 8080 .
			input:  []byte{184, 101, 1, 32, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 41, 4, 208, 0, 0, 0, 0, 0, 12, 0, 10, 0, 8, 109, 145, 230, 150, 58, 147, 7, 196},
			offset: wire.HeaderSize,
			question: &message.Question{
				Name:  domain.Domain{{}},
				Type:  message.QueryType(record.TypeA),
				Class: message.QueryClass(record.ClassIN),
			},
			size: new(uint(5)),
			err:  nil,
		},
		{
			// dig @127.0.0.1 -p 8080 -c CH -t NS foo.local.
			input:  []byte{68, 29, 1, 32, 0, 1, 0, 0, 0, 0, 0, 1, 3, 102, 111, 111, 5, 108, 111, 99, 97, 108, 0, 0, 2, 0, 3, 0, 0, 41, 4, 208, 0, 0, 0, 0, 0, 12, 0, 10, 0, 8, 124, 39, 56, 53, 31, 181, 54, 190},
			offset: wire.HeaderSize,
			question: &message.Question{
				Name: domain.Domain{
					{'f', 'o', 'o'},
					{'l', 'o', 'c', 'a', 'l'},
					{},
				},
				Type:  message.QueryType(record.TypeNS),
				Class: message.QueryClass(record.ClassCH),
			},
			size: new(uint(15)),
			err:  nil,
		},
	}

	for _, tc := range tests {
		parser := wire.NewParser(tc.input, wire.WithOffset(tc.offset))
		question, err := parser.ParseQuestion()

		assert.Equal(t, tc.err, err)

		if tc.question != nil {
			assert.Equal(t, *tc.question, question)
			assert.Equal(t, *tc.size, parser.Offset()-tc.offset)
		}
	}
}

func testParseResourceRecord(t *testing.T) {
	tests := []struct {
		input  []byte
		offset uint
		rr     *record.Record
		size   *uint
		err    error
	}{
		{
			input:  []byte{0x75, 0x0d, 0x85, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x03, 0x66, 0x6f, 0x6f, 0x05, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x06, 0x00, 0x01, 0x00, 0x01, 0x51, 0x80, 0x00, 0x40, 0x01, 0x61, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0x05, 0x6e, 0x73, 0x74, 0x6c, 0x64, 0x0c, 0x76, 0x65, 0x72, 0x69, 0x73, 0x69, 0x67, 0x6e, 0x2d, 0x67, 0x72, 0x73, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x78, 0xc3, 0xb0, 0xcd, 0x00, 0x00, 0x07, 0x08, 0x00, 0x00, 0x03, 0x84, 0x00, 0x09, 0x3a, 0x80, 0x00, 0x01, 0x51, 0x80, 0x00, 0x00, 0x29, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			offset: 27,
			rr: &record.Record{
				Name:  domain.Domain{{}},
				Type:  record.TypeSOA,
				Class: record.ClassIN,
				TTL:   86400,
				Data: []byte{
					0x01, 'a', 0x0c, 'r', 'o', 'o', 't', '-', 's', 'e', 'r', 'v', 'e', 'r', 's', 0x03, 'n', 'e', 't', 0x00,
					0x05, 'n', 's', 't', 'l', 'd', 0x0c, 'v', 'e', 'r', 'i', 's', 'i', 'g', 'n', '-', 'g', 'r', 's', 0x03, 'c', 'o', 'm', 0x00,
					0x78, 0xc3, 0xb0, 0xcd,
					0x00, 0x00, 0x07, 0x08,
					0x00, 0x00, 0x03, 0x84,
					0x00, 0x09, 0x3a, 0x80,
					0x00, 0x01, 0x51, 0x80,
				},
			},
			size: new(uint(75)),
			err:  nil,
		},
		{
			input:  []byte{0x90, 0xdf, 0x81, 0x80, 0x00, 0x01, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x05, 0x64, 0x64, 0x6e, 0x65, 0x74, 0x03, 0x6f, 0x72, 0x67, 0x00, 0x00, 0x01, 0x00, 0x01, 0xc0, 0x0c, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x2c, 0x00, 0x04, 0x68, 0x12, 0x0a, 0x2c, 0xc0, 0x0c, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x2c, 0x00, 0x04, 0x68, 0x12, 0x0b, 0x2c, 0x00, 0x00, 0x29, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			offset: 27,
			rr: &record.Record{
				Name:  domain.Domain{{'d', 'd', 'n', 'e', 't'}, {'o', 'r', 'g'}, {}},
				Type:  record.TypeA,
				Class: record.ClassIN,
				TTL:   300,
				Data:  []byte{104, 18, 10, 44},
			},
			size: new(uint(16)),
			err:  nil,
		},
	}

	for _, tc := range tests {
		parser := wire.NewParser(tc.input, wire.WithOffset(tc.offset))
		rr, err := parser.ParseResourceRecord()

		assert.Equal(t, tc.err, err)

		if tc.rr != nil {
			assert.Equal(t, *tc.rr, rr)
			assert.Equal(t, *tc.size, parser.Offset()-tc.offset)
		}
	}
}

func testParseMessage(t *testing.T) {
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
		parser := wire.NewParser(tc.input)
		message, err := parser.ParseMessage()

		assert.Equal(t, tc.err, err)

		if tc.message != nil {
			assert.Equal(t, *tc.message, message)
		}
	}
}
