package wire

import (
	"errors"

	"github.com/MilkeeyCat/bdns/message"
	"github.com/MilkeeyCat/bdns/record"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrShortMessage   = errors.New("short message")
)

func ParseMessage(buf []byte) (message.Message, error) {
	if len(buf) < HeaderSize {
		return message.Message{}, ErrShortMessage
	}

	header, err := ParseHeader([HeaderSize]byte(buf[:HeaderSize]))
	if err != nil {
		return message.Message{}, err
	}

	questions := make([]message.Question, header.QDCount)
	offset := uint(HeaderSize)

	for i := range header.QDCount {
		question, size, err := ParseQuestion(buf, offset)
		if err != nil {
			return message.Message{}, err
		}

		questions[i] = question
		offset += size
	}

	answer, size, err := parseRRs(buf, header.ANCount, offset)
	if err != nil {
		return message.Message{}, err
	}

	offset += size

	authority, size, err := parseRRs(buf, header.NSCount, offset)
	if err != nil {
		return message.Message{}, err
	}

	offset += size

	additional, size, err := parseRRs(buf, header.ARCount, offset)
	if err != nil {
		return message.Message{}, err
	}

	offset += size

	if len(buf[offset:]) > 0 {
		return message.Message{}, ErrInvalidMessage
	}

	return message.Message{
		Header:     header.Header,
		Question:   questions,
		Answer:     answer,
		Authority:  authority,
		Additional: additional,
	}, nil
}

func parseRRs(buf []byte, n uint16, offset uint) ([]record.Record, uint, error) {
	rrs := make([]record.Record, n)
	var size uint

	for i := range n {
		rr, rrSize, err := ParseResourceRecord(buf, offset+size)
		if err != nil {
			return nil, 0, err
		}

		rrs[i] = rr
		size += rrSize
	}

	return rrs, size, nil
}
