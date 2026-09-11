package wire_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/message"
	"github.com/MilkeeyCat/bdns/record"
	"github.com/MilkeeyCat/bdns/wire"
)

func TestEncoder(t *testing.T) {
	t.Run("domain", testEncodeDomain)
	t.Run("question", testEncodeQuestion)
	t.Run("resource", testEncodeResourceRecord)
}

func testEncodeDomain(t *testing.T) {
	tests := []struct {
		domains []string
		result  []byte
	}{
		{
			domains: []string{"f.isi."},
			result:  []byte{0x01, 'f', 0x03, 'i', 's', 'i', 0x00},
		},
		{
			domains: []string{"m.isi.arpa.", "foo.m.isi.arpa."},
			result: []byte{
				0x01, 'm', 0x03, 'i', 's', 'i', 0x04, 'a', 'r', 'p', 'a', 0x00,
				0x03, 'f', 'o', 'o', 0xc0, 0x00,
			},
		},
		{
			domains: []string{
				"qux.com.",
				"baz.qux.com.",
				"bar.baz.qux.com.",
			},
			result: []byte{
				0x03, 'q', 'u', 'x', 0x03, 'c', 'o', 'm', 0x00,
				0x03, 'b', 'a', 'z', 0xc0, 0x00,
				0x03, 'b', 'a', 'r', 0xc0, 0x09,
			},
		},
	}

	for _, tc := range tests {
		var encoder wire.Encoder

		for _, strDomain := range tc.domains {
			domain := domain.NewDomainFromString(strDomain)

			assert.NoError(t, encoder.EncodeDomain(domain))
		}

		assert.Equal(t, tc.result, encoder.Bytes())
	}
}

func testEncodeQuestion(t *testing.T) {
	tests := []struct {
		question message.Question
		result   []byte
		err      error
	}{
		{
			question: message.Question{
				Name:  domain.Domain{{}},
				Type:  message.QueryType(record.TypeA),
				Class: message.QueryClass(record.ClassIN),
			},
			result: []byte{0x00, 0x00, 0x01, 0x00, 0x01},
			err:    nil,
		},
		{
			question: message.Question{
				Name: domain.Domain{
					{'f', 'o', 'o'},
					{'l', 'o', 'c', 'a', 'l'},
					{},
				},
				Type:  message.QueryType(record.TypeNS),
				Class: message.QueryClass(record.ClassCH),
			},
			result: []byte{0x03, 0x66, 0x6f, 0x6f, 0x05, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x00, 0x00, 0x02, 0x00, 0x03},
			err:    nil,
		},
	}

	for _, tc := range tests {
		var encoder wire.Encoder

		err := encoder.EncodeQuestion(tc.question)

		assert.Equal(t, tc.err, err)
		assert.Equal(t, tc.result, encoder.Bytes())
	}
}

func testEncodeResourceRecord(t *testing.T) {
	tests := []struct {
		rr     record.Record
		result []byte
		err    error
	}{
		{
			rr: record.Record{
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
			result: []byte{0x00, 0x00, 0x06, 0x00, 0x01, 0x00, 0x01, 0x51, 0x80, 0x00, 0x40, 0x01, 0x61, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0x05, 0x6e, 0x73, 0x74, 0x6c, 0x64, 0x0c, 0x76, 0x65, 0x72, 0x69, 0x73, 0x69, 0x67, 0x6e, 0x2d, 0x67, 0x72, 0x73, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x78, 0xc3, 0xb0, 0xcd, 0x00, 0x00, 0x07, 0x08, 0x00, 0x00, 0x03, 0x84, 0x00, 0x09, 0x3a, 0x80, 0x00, 0x01, 0x51, 0x80},
			err:    nil,
		},
		{
			rr: record.Record{
				Name:  domain.Domain{{'d', 'd', 'n', 'e', 't'}, {'o', 'r', 'g'}, {}},
				Type:  record.TypeA,
				Class: record.ClassIN,
				TTL:   300,
				Data:  []byte{104, 18, 10, 44},
			},
			result: []byte{0x05, 0x64, 0x64, 0x6e, 0x65, 0x74, 0x03, 0x6f, 0x72, 0x67, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x2c, 0x00, 0x04, 0x68, 0x12, 0x0a, 0x2c},
			err:    nil,
		},
	}

	for _, tc := range tests {
		var encoder wire.Encoder

		err := encoder.EncodeResourceRecord(tc.rr)

		assert.Equal(t, tc.err, err)
		assert.Equal(t, tc.result, encoder.Bytes())
	}
}
