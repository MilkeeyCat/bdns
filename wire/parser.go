package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net/netip"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/message"
	"github.com/MilkeeyCat/bdns/record"
)

var ErrInvalidMessage = errors.New("invalid message")

type Parser struct {
	r *bytes.Reader
}

type ParserOption func(*Parser)

func WithOffset(offset uint) ParserOption {
	return func(p *Parser) {
		if _, err := p.r.Seek(int64(offset), io.SeekStart); err != nil {
			panic(err)
		}
	}
}

func NewParser(buf []byte, options ...ParserOption) *Parser {
	p := &Parser{
		r: bytes.NewReader(buf),
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

func (p *Parser) Offset() uint {
	offset, err := p.r.Seek(0, io.SeekCurrent)
	if err != nil {
		panic(err)
	}

	return uint(offset)
}

func (p *Parser) ParseDomain() (domain.Domain, error) {
	return p.parseDomain(nil, make(map[uint16]struct{}))
}

func (p *Parser) parseDomain(
	domain domain.Domain,
	offsets map[uint16]struct{},
) (domain.Domain, error) {
	for {
		length, err := p.r.ReadByte()
		if err != nil {
			return nil, err
		}

		switch length & 0b1100_0000 {
		case 0b1100_0000:
			b, err := p.r.ReadByte()
			if err != nil {
				return nil, err
			}

			offset := (uint16(length&0b0011_1111) << 8) | uint16(b)

			if _, ok := offsets[offset]; ok {
				return nil, ErrInvalidMessage
			}

			oldOffset := p.Offset()

			if _, err := p.r.Seek(int64(offset), io.SeekStart); err != nil {
				return nil, err
			}

			offsets[offset] = struct{}{}

			domain, err := p.parseDomain(domain, offsets)
			if err != nil {
				return nil, err
			}

			if _, err := p.r.Seek(int64(oldOffset), io.SeekStart); err != nil {
				return nil, err
			}

			return domain, err

		case 0:
			buf := make([]byte, length)

			if _, err := io.ReadFull(p.r, buf[:]); err != nil {
				return nil, err
			}

			domain = append(domain, buf)

			if length == 0 {
				return domain, nil
			}

		default:
			return nil, ErrInvalidMessage
		}
	}
}

func (p *Parser) ParseQuestion() (message.Question, error) {
	domain, err := p.ParseDomain()
	if err != nil {
		return message.Question{}, err
	}

	const staticDataSize = 2 + 2

	var buf [staticDataSize]byte

	if _, err := io.ReadFull(p.r, buf[:]); err != nil {
		return message.Question{}, err
	}

	queryType, err := parseQueryType(binary.BigEndian.Uint16(buf[0:]))
	if err != nil {
		return message.Question{}, err
	}

	queryClass, err := parseQueryClass(binary.BigEndian.Uint16(buf[2:]))
	if err != nil {
		return message.Question{}, err
	}

	return message.Question{
		Name:  domain,
		Type:  queryType,
		Class: queryClass,
	}, nil
}

func (p *Parser) parseCharString() (string, error) {
	size, err := p.r.ReadByte()
	if err != nil {
		return "", err
	}

	buf := make([]byte, size)

	if _, err := io.ReadFull(p.r, buf); err != nil {
		return "", err
	}

	return string(buf), nil
}

func (p *Parser) parseUint32() (uint32, error) {
	var buf [4]byte

	if _, err := io.ReadFull(p.r, buf[:]); err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(buf[:]), nil
}

func (p *Parser) parseRRData(
	ty record.Type,
	class record.Class,
	length uint16,
) (record.Data, error) {
	offset := p.Offset()
	var data record.Data

	switch {
	case ty == record.TypeCNAME:
		cname, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		data = record.CNAMEData{
			CNAME: cname,
		}

	case ty == record.TypeHINFO:
		cpu, err := p.parseCharString()
		if err != nil {
			return nil, err
		}

		os, err := p.parseCharString()
		if err != nil {
			return nil, err
		}

		data = record.HINFOData{
			CPU: cpu,
			OS:  os,
		}

	case ty == record.TypeMB:
		madName, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		data = record.MBData{
			MADName: madName,
		}

	case ty == record.TypeMG:
		mgmName, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		data = record.MGData{
			MGMName: mgmName,
		}

	case ty == record.TypeMINFO:
		rMailbx, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		eMailbx, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		data = record.MINFOData{
			RMailbx: rMailbx,
			EMailbx: eMailbx,
		}

	case ty == record.TypeMR:
		newName, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		data = record.MRData{
			NewName: newName,
		}

	case ty == record.TypeMX:
		var buf [2]byte

		if _, err := io.ReadFull(p.r, buf[:]); err != nil {
			return nil, err
		}

		exchange, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		data = record.MXData{
			Preference: binary.BigEndian.Uint16(buf[:]),
			Exchange:   exchange,
		}

	case ty == record.TypeNULL:
		buf := make([]byte, length)

		if _, err := io.ReadFull(p.r, buf); err != nil {
			return nil, err
		}

		data = record.NULLData{
			Data: buf,
		}

	case ty == record.TypeNS:
		nsdName, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		data = record.NSData{
			NSDName: nsdName,
		}

	case ty == record.TypePTR:
		domain, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		data = record.PTRData{
			PTRDName: domain,
		}

	case ty == record.TypeSOA:
		mName, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		rName, err := p.ParseDomain()
		if err != nil {
			return nil, err
		}

		serial, err := p.parseUint32()
		if err != nil {
			return nil, err
		}

		refresh, err := p.parseUint32()
		if err != nil {
			return nil, err
		}

		retry, err := p.parseUint32()
		if err != nil {
			return nil, err
		}

		expire, err := p.parseUint32()
		if err != nil {
			return nil, err
		}

		minimum, err := p.parseUint32()
		if err != nil {
			return nil, err
		}

		data = record.SOAData{
			MName:   mName,
			RName:   rName,
			Serial:  serial,
			Refresh: refresh,
			Retry:   retry,
			Expire:  expire,
			Minimum: minimum,
		}

	case ty == record.TypeTXT:
		var strs []string

		for p.Offset() < offset+uint(length) {
			str, err := p.parseCharString()
			if err != nil {
				return nil, err
			}

			strs = append(strs, str)
		}

		data = record.TXTData{
			Data: strs,
		}

	case ty == record.TypeA && class == record.ClassIN:
		var buf [4]byte

		if _, err := io.ReadFull(p.r, buf[:]); err != nil {
			return nil, err
		}

		data = record.AData{
			Address: netip.AddrFrom4(buf),
		}

	case ty == record.TypeWKS && class == record.ClassIN:
		var buf [4]byte

		if _, err := io.ReadFull(p.r, buf[:]); err != nil {
			return nil, err
		}

		protocol, err := p.r.ReadByte()
		if err != nil {
			return nil, err
		}

		bitmap := make([]byte, length-4-1)

		if _, err := io.ReadFull(p.r, bitmap); err != nil {
			return nil, err
		}

		data = record.WKSData{
			Address:  netip.AddrFrom4(buf),
			Protocol: protocol,
			Bitmap:   bitmap,
		}

	default:
		return nil, record.ErrUnsupportedRecordData
	}

	if p.Offset() != offset+uint(length) {
		return nil, ErrInvalidMessage
	}

	return data, nil
}

func (p *Parser) ParseResourceRecord() (record.Record, error) {
	domain, err := p.ParseDomain()
	if err != nil {
		return record.Record{}, err
	}

	const staticDataSize = 2 + 2 + 4 + 2

	var buf [staticDataSize]byte

	if _, err := io.ReadFull(p.r, buf[:]); err != nil {
		return record.Record{}, err
	}

	ty, err := parseType(binary.BigEndian.Uint16(buf[0:]))
	if err != nil {
		return record.Record{}, err
	}

	class, err := parseClass(binary.BigEndian.Uint16(buf[2:]))
	if err != nil {
		return record.Record{}, err
	}

	ttl := binary.BigEndian.Uint32(buf[4:])
	rdLength := binary.BigEndian.Uint16(buf[8:])
	data, err := p.parseRRData(ty, class, rdLength)
	if err != nil {
		return record.Record{}, err
	}

	return record.Record{
		Name:  domain,
		Type:  ty,
		Class: class,
		TTL:   ttl,
		Data:  data,
	}, nil
}

func (p *Parser) ParseMessage() (message.Message, error) {
	var buf [HeaderSize]byte

	if _, err := io.ReadFull(p.r, buf[:]); err != nil {
		return message.Message{}, err
	}

	header, err := ParseHeader(buf)
	if err != nil {
		return message.Message{}, err
	}

	questions := make([]message.Question, header.QDCount)

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

	if p.r.Len() > 0 {
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
