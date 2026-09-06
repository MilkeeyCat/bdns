package wire_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/message"
	"github.com/MilkeeyCat/bdns/record"
	"github.com/MilkeeyCat/bdns/wire"
)

func TestQuestion(t *testing.T) {
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
		question, size, err := wire.ParseQuestion(tc.input, tc.offset)

		if tc.question != nil {
			assert.Equal(t, *tc.question, question)
			assert.Equal(t, *tc.size, size)
		}

		assert.Equal(t, tc.err, err)
	}
}
