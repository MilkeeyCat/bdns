package wire

import (
	"encoding/binary"

	"github.com/MilkeeyCat/bdns/record"
)

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

func ParseResourceRecord(buf []byte, offset uint) (record.Record, uint, error) {
	domain, domainSize, err := ParseDomain(buf, offset)
	if err != nil {
		return record.Record{}, 0, err
	}

	buf = buf[offset+uint(domainSize):]

	const staticDataSize = 2 + 2 + 4 + 2

	if len(buf) < staticDataSize {
		return record.Record{}, 0, ErrShortMessage
	}

	ty, err := parseType(binary.BigEndian.Uint16(buf))
	if err != nil {
		return record.Record{}, 0, err
	}

	class, err := parseClass(binary.BigEndian.Uint16(buf[2:]))
	if err != nil {
		return record.Record{}, 0, err
	}

	ttl := binary.BigEndian.Uint32(buf[4:])
	rdLength := binary.BigEndian.Uint16(buf[8:])

	if len(buf) < int(rdLength)+staticDataSize {
		return record.Record{}, 0, ErrShortMessage
	}

	size := uint(domainSize) + staticDataSize + uint(rdLength)

	return record.Record{
		Name:  domain,
		Type:  ty,
		Class: class,
		TTL:   ttl,
		Data:  buf[staticDataSize : rdLength+staticDataSize],
	}, size, nil
}
