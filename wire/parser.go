package wire

import (
	"encoding/binary"
	"errors"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/message"
	"github.com/MilkeeyCat/bdns/record"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrShortMessage   = errors.New("short message")
)

type Parser struct {
	buf    []byte
	offset uint
}

type Option func(*Parser)

func WithOffset(cursor uint) Option {
	return func(p *Parser) {
		p.offset = cursor
	}
}

func NewParser(buf []byte, options ...Option) *Parser {
	p := &Parser{
		buf:    buf,
		offset: 0,
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

func (p *Parser) Offset() uint {
	return p.offset
}

func (p *Parser) ParseDomain() (domain.Domain, error) {
	result, size, err := p.parseDomain(nil, p.offset, 0, make(map[uint16]struct{}))
	if err != nil {
		return domain.Domain{}, err
	}

	p.offset += uint(size)

	return result, nil
}

func (p *Parser) parseDomain(
	domain domain.Domain,
	offset uint,
	size uint8,
	offsets map[uint16]struct{},
) (domain.Domain, uint8, error) {
	for {
		if len(p.buf[offset:]) < 1 {
			return nil, 0, ErrShortMessage
		}

		length := p.buf[offset]

		switch length & 0b1100_0000 {
		case 0b1100_0000:
			if len(p.buf[offset:]) < 2 {
				return nil, 0, ErrShortMessage
			}

			offset := (uint16(length&0b0011_1111) << 8) | uint16(p.buf[offset+1])

			if _, ok := offsets[offset]; ok {
				return nil, 0, ErrInvalidMessage
			}

			if len(p.buf) <= int(offset) {
				return nil, 0, ErrInvalidMessage
			}

			offsets[offset] = struct{}{}

			domain, _, err := p.parseDomain(domain, uint(offset), size, offsets)
			if err != nil {
				return nil, 0, err
			}

			return domain, size + 2, err

		case 0:
			if len(p.buf[offset:]) < int(length) {
				return nil, 0, ErrShortMessage
			}

			domain = append(domain, p.buf[offset+1:offset+uint(length)+1])
			offset += uint(length) + 1
			size += length + 1

			if length == 0 {
				return domain, size, nil
			}

		default:
			return nil, 0, ErrInvalidMessage
		}
	}
}

func (p *Parser) ParseQuestion() (message.Question, error) {
	domain, err := p.ParseDomain()
	if err != nil {
		return message.Question{}, err
	}

	buf := p.buf[p.offset:]

	const staticDataSize = 2 + 2

	if len(buf) < staticDataSize {
		return message.Question{}, ErrShortMessage
	}

	queryType, err := parseQueryType(binary.BigEndian.Uint16(buf[0:]))
	if err != nil {
		return message.Question{}, err
	}

	queryClass, err := parseQueryClass(binary.BigEndian.Uint16(buf[2:]))
	if err != nil {
		return message.Question{}, err
	}

	p.offset += staticDataSize

	return message.Question{
		Name:  domain,
		Type:  queryType,
		Class: queryClass,
	}, nil
}

func (p *Parser) ParseResourceRecord() (record.Record, error) {
	domain, err := p.ParseDomain()
	if err != nil {
		return record.Record{}, err
	}

	buf := p.buf[p.offset:]

	const staticDataSize = 2 + 2 + 4 + 2

	if len(buf) < staticDataSize {
		return record.Record{}, ErrShortMessage
	}

	ty, err := parseType(binary.BigEndian.Uint16(buf))
	if err != nil {
		return record.Record{}, err
	}

	class, err := parseClass(binary.BigEndian.Uint16(buf[2:]))
	if err != nil {
		return record.Record{}, err
	}

	ttl := binary.BigEndian.Uint32(buf[4:])
	rdLength := binary.BigEndian.Uint16(buf[8:])

	if len(buf) < int(rdLength)+staticDataSize {
		return record.Record{}, ErrShortMessage
	}

	p.offset += staticDataSize + uint(rdLength)

	return record.Record{
		Name:  domain,
		Type:  ty,
		Class: class,
		TTL:   ttl,
		Data:  buf[staticDataSize : rdLength+staticDataSize],
	}, nil
}

func (p *Parser) ParseMessage() (message.Message, error) {
	if len(p.buf) < HeaderSize {
		return message.Message{}, ErrShortMessage
	}

	header, err := ParseHeader([HeaderSize]byte(p.buf[:HeaderSize]))
	if err != nil {
		return message.Message{}, err
	}

	questions := make([]message.Question, header.QDCount)

	p.offset = HeaderSize

	for i := range header.QDCount {
		question, err := p.ParseQuestion()
		if err != nil {
			return message.Message{}, err
		}

		questions[i] = question
	}

	answer, err := p.parseRRs(header.ANCount)
	if err != nil {
		return message.Message{}, err
	}

	authority, err := p.parseRRs(header.NSCount)
	if err != nil {
		return message.Message{}, err
	}

	additional, err := p.parseRRs(header.ARCount)
	if err != nil {
		return message.Message{}, err
	}

	if len(p.buf[p.offset:]) > 0 {
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

func (p *Parser) parseRRs(n uint16) ([]record.Record, error) {
	rrs := make([]record.Record, n)

	for i := range n {
		rr, err := p.ParseResourceRecord()
		if err != nil {
			return nil, err
		}

		rrs[i] = rr
	}

	return rrs, nil
}

func parseType(code uint16) (record.Type, error) {
	switch code {
	case 1:
		return record.TypeA, nil
	case 2:
		return record.TypeNS, nil
	case 3:
		return record.TypeMD, nil
	case 4:
		return record.TypeMF, nil
	case 5:
		return record.TypeCNAME, nil
	case 6:
		return record.TypeSOA, nil
	case 7:
		return record.TypeMB, nil
	case 8:
		return record.TypeMG, nil
	case 9:
		return record.TypeMR, nil
	case 10:
		return record.TypeNULL, nil
	case 11:
		return record.TypeWKS, nil
	case 12:
		return record.TypePTR, nil
	case 13:
		return record.TypeHINFO, nil
	case 14:
		return record.TypeMINFO, nil
	case 15:
		return record.TypeMX, nil
	case 16:
		return record.TypeTXT, nil
	default:
		return 0, ErrInvalidMessage
	}
}

func parseClass(code uint16) (record.Class, error) {
	switch code {
	case 1:
		return record.ClassIN, nil
	case 2:
		return record.ClassCS, nil
	case 3:
		return record.ClassCH, nil
	case 4:
		return record.ClassHS, nil
	default:
		return 0, ErrInvalidMessage
	}
}

func parseQueryType(code uint16) (message.QueryType, error) {
	switch code {
	case 252:
		return message.QueryTypeAXFR, nil
	case 253:
		return message.QueryTypeMAILB, nil
	case 254:
		return message.QueryTypeMAILA, nil
	case 255:
		return message.QueryTypeAll, nil
	default:
		ty, err := parseType(code)
		if err != nil {
			return 0, err
		}

		return message.QueryType(ty), nil
	}
}

func parseQueryClass(code uint16) (message.QueryClass, error) {
	switch code {
	case 255:
		return message.QueryClassAny, nil
	default:
		class, err := parseClass(code)
		if err != nil {
			return 0, err
		}

		return message.QueryClass(class), nil
	}
}
