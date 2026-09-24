package masterfile

import (
	"errors"
	"fmt"
	"io"
	"net/netip"
	"strconv"
	"strings"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/record"
)

var (
	errNoOrigin                = errors.New("origin is not set")
	errEmptyNonRootLabel       = errors.New("empty non-root label")
	errFailedToParseRecordType = errors.New("failed to parse RR type")
	errNoOwnerToInherit        = errors.New("no owner to inherit")
	errNoClassToInherit        = errors.New("no class to inherit")
	errNoTTLToInherit          = errors.New("no TTL to inherit")
	errDuplicateClass          = errors.New("duplicate class")
	errDuplicateTTL            = errors.New("duplicate TTL")
	errRelativeDomainOrigin    = errors.New("relative domain origin")
)

type unexpectedTokenTypeError struct {
	expected []tokenType
	actual   tokenType
}

func (e unexpectedTokenTypeError) Error() string {
	var buf strings.Builder

	_, _ = buf.WriteString("unexpected token type: expected ")

	if len(e.expected) > 1 {
		buf.WriteString("one of ")
	}

	for i, tt := range e.expected {
		_, _ = buf.WriteString(tt.String())

		if i < len(e.expected)-1 {
			_, _ = buf.WriteString(", ")
		}
	}

	_, _ = buf.WriteString("; actual ")
	_, _ = buf.WriteString(e.actual.String())

	return buf.String()
}

type parser struct {
	l         lexer
	curToken  token
	peekToken token

	origin domain.Domain
	domain domain.Domain
	class  *record.Class
	ttl    *uint32
}

func newParser(rd io.Reader) (parser, error) {
	p := parser{
		l: newLexer(rd),
	}

	for range 2 {
		if err := p.bump(); err != nil {
			return p, err
		}
	}

	return p, nil
}

func (p *parser) bump() error {
	token, err := p.l.next()
	if err != nil {
		return err
	}

	p.curToken = p.peekToken
	p.peekToken = token

	return nil
}

func (p *parser) currentStringRaw() (string, error) {
	if p.curToken.tt != tokenTypeString {
		return "", unexpectedTokenTypeError{
			expected: []tokenType{tokenTypeString},
			actual:   p.curToken.tt,
		}
	}

	return p.curToken.value, nil
}

func (p *parser) parseStringRaw() (string, error) {
	if p.curToken.tt != tokenTypeString {
		return "", unexpectedTokenTypeError{
			expected: []tokenType{tokenTypeString},
			actual:   p.curToken.tt,
		}
	}

	str := p.curToken.value

	if err := p.bump(); err != nil {
		return "", err
	}

	return str, nil
}

func (p *parser) parseString() (string, error) {
	str, err := p.parseStringRaw()
	if err != nil {
		return "", err
	}

	return canonicalizeString(str), nil
}

func (p *parser) parseType() (record.Type, bool) {
	var ty record.Type
	str, err := p.currentStringRaw()
	if err != nil {
		return ty, false
	}

	switch str {
	case "A":
		ty = record.TypeA
	case "NS":
		ty = record.TypeNS
	case "CNAME":
		ty = record.TypeCNAME
	case "SOA":
		ty = record.TypeSOA
	case "MB":
		ty = record.TypeMB
	case "MG":
		ty = record.TypeMG
	case "MR":
		ty = record.TypeMR
	case "NULL":
		ty = record.TypeNULL
	case "WKS":
		ty = record.TypeWKS
	case "PTR":
		ty = record.TypePTR
	case "HINFO":
		ty = record.TypeHINFO
	case "MINFO":
		ty = record.TypeMINFO
	case "MX":
		ty = record.TypeMX
	case "TXT":
		ty = record.TypeTXT
	default:
		return ty, false
	}

	if err := p.bump(); err != nil {
		return ty, false
	}

	return ty, true
}

func (p *parser) parseClass() (record.Class, bool) {
	var class record.Class
	str, err := p.currentStringRaw()
	if err != nil {
		return class, false
	}

	switch str {
	case "IN":
		class = record.ClassIN
	case "CS":
		class = record.ClassCS
	case "CH":
		class = record.ClassCH
	case "HS":
		class = record.ClassHS
	default:
		return class, false
	}

	if err := p.bump(); err != nil {
		return class, false
	}

	return class, true
}

func (p *parser) parseTTL() (uint32, bool) {
	str, err := p.currentStringRaw()
	if err != nil {
		return 0, false
	}

	ttl, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, false
	}

	if err := p.bump(); err != nil {
		return 0, false
	}

	return uint32(ttl), true
}

func (p *parser) parseUint32() (uint32, error) {
	str, err := p.parseStringRaw()
	if err != nil {
		return 0, err
	}

	ttl, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(ttl), nil
}

func (p *parser) parseDomain() (domain.Domain, error) {
	str, err := p.parseStringRaw()
	if err != nil {
		return nil, err
	}

	switch str {
	case "@":
		if p.origin == nil {
			return nil, errNoOrigin
		}

		return p.origin, nil

	case ".":
		return domain.Domain{{}}, nil
	}

	runes := []rune(str)
	n := len(runes)
	start := 0
	var labels []domain.Label

	// this doesn't look pretty but oh well wcyd
	for i := 0; i <= n && n > 0; i++ {
		if i < n && runes[i] == '\\' {
			if runes[i+1] >= '0' && runes[i+1] <= '9' {
				i += 3
			} else {
				i++
			}
		} else if i == n || runes[i] == '.' {
			if i-start == 0 && i != n {
				return nil, errEmptyNonRootLabel
			}

			labels = append(
				labels,
				domain.Label(canonicalizeString(string(runes[start:i]))),
			)

			start = i + 1
		}
	}

	if len(labels) == 0 || len(labels) > 0 && !labels[len(labels)-1].IsRoot() {
		if p.origin == nil {
			return nil, errNoOrigin
		}

		labels = append(labels, p.origin...)
	}

	return labels, nil
}

func (p *parser) parseAddr() (netip.Addr, error) {
	str, err := p.parseString()
	if err != nil {
		return netip.Addr{}, err
	}

	addr, err := netip.ParseAddr(str)
	if err != nil {
		return netip.Addr{}, err
	}

	if !addr.Is4() {
		return netip.Addr{}, errors.New("unable to parse IP")
	}

	return addr, nil
}

func (p *parser) parseRRData(ty record.Type, class record.Class) (record.Data, error) {
	switch {
	case ty == record.TypeCNAME:
		cname, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		return record.CNAMEData{
			CNAME: cname,
		}, nil

	case ty == record.TypeHINFO:
		cpu, err := p.parseString()
		if err != nil {
			return nil, err
		}

		os, err := p.parseString()
		if err != nil {
			return nil, err
		}

		return record.HINFOData{
			CPU: cpu,
			OS:  os,
		}, nil

	case ty == record.TypeMB:
		madName, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		return record.MBData{
			MADName: madName,
		}, nil

	case ty == record.TypeMG:
		mgmName, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		return record.MGData{
			MGMName: mgmName,
		}, nil

	case ty == record.TypeMINFO:
		rMailbx, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		eMailbx, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		return record.MINFOData{
			RMailbx: rMailbx,
			EMailbx: eMailbx,
		}, nil

	case ty == record.TypeMR:
		newName, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		return record.MRData{
			NewName: newName,
		}, nil

	case ty == record.TypeMX:
		preferenceStr, err := p.parseString()
		if err != nil {
			return nil, err
		}

		preference, err := strconv.ParseUint(preferenceStr, 10, 16)
		if err != nil {
			return nil, err
		}

		exchange, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		return record.MXData{
			Preference: uint16(preference),
			Exchange:   exchange,
		}, nil

	case ty == record.TypeNULL:
		str, err := p.parseString()
		if err != nil {
			return nil, err
		}

		return record.NULLData{
			Data: []byte(str),
		}, nil

	case ty == record.TypeNS:
		nsdName, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		return record.NSData{
			NSDName: nsdName,
		}, nil

	case ty == record.TypePTR:
		domain, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		return record.PTRData{
			PTRDName: domain,
		}, nil

	case ty == record.TypeSOA:
		mName, err := p.parseDomain()
		if err != nil {
			return nil, err
		}

		rName, err := p.parseDomain()
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

		return record.SOAData{
			MName:   mName,
			RName:   rName,
			Serial:  serial,
			Refresh: refresh,
			Retry:   retry,
			Expire:  expire,
			Minimum: minimum,
		}, nil

	case ty == record.TypeTXT:
		var strs []string

		for p.curToken.tt != tokenTypeNewline && p.curToken.tt != tokenTypeEOF {
			str, err := p.parseString()
			if err != nil {
				return nil, err
			}

			strs = append(strs, str)
		}

		return record.TXTData{
			Data: strs,
		}, nil

	case ty == record.TypeA && class == record.ClassIN:
		addr, err := p.parseAddr()
		if err != nil {
			return nil, err
		}

		return record.AData{
			Address: addr,
		}, nil

	case ty == record.TypeWKS && class == record.ClassIN:
		addr, err := p.parseAddr()
		if err != nil {
			return nil, err
		}

		str, err := p.parseString()
		if err != nil {
			return nil, err
		}

		protocol, err := strconv.ParseUint(str, 10, 8)
		if err != nil {
			return nil, err
		}

		bitmap, err := p.parseString()
		if err != nil {
			return nil, err
		}

		return record.WKSData{
			Address:  addr,
			Protocol: uint8(protocol),
			Bitmap:   []byte(bitmap),
		}, nil

	default:
		return nil, record.ErrUnsupportedRecordData
	}
}

func (p *parser) parseRR() (record.Record, error) {
	var domain domain.Domain

	switch p.curToken.tt {
	case tokenTypeIndent:
		if p.domain == nil {
			return record.Record{}, errNoOwnerToInherit
		}

		domain = p.domain

		if err := p.bump(); err != nil {
			return record.Record{}, err
		}

	case tokenTypeString:
		var err error

		if domain, err = p.parseDomain(); err != nil {
			return record.Record{}, err
		}

		p.domain = domain

	default:
		panic(fmt.Sprintf("unexpected curToken: %v", p.curToken))
	}

	var (
		ty    *record.Type
		class *record.Class
		ttl   *uint32
	)

	for range 3 {
		if parsedType, ok := p.parseType(); ok {
			ty = &parsedType

			break
		}

		if parsedClass, ok := p.parseClass(); ok {
			if class != nil {
				return record.Record{}, errDuplicateClass
			}

			class = &parsedClass

			continue
		}

		if parsedTTL, ok := p.parseTTL(); ok {
			if ttl != nil {
				return record.Record{}, errDuplicateTTL
			}

			ttl = &parsedTTL

			continue
		}
	}

	if ty == nil {
		return record.Record{}, errFailedToParseRecordType
	}

	if class == nil {
		if p.class == nil {
			return record.Record{}, errNoClassToInherit
		}

		class = p.class
	} else {
		p.class = class
	}

	if ttl == nil {
		if p.ttl == nil {
			return record.Record{}, errNoTTLToInherit
		}

		ttl = p.ttl
	} else {
		p.ttl = ttl
	}

	data, err := p.parseRRData(*ty, *class)
	if err != nil {
		return record.Record{}, err
	}

	if p.curToken.tt != tokenTypeNewline && p.curToken.tt != tokenTypeEOF {
		return record.Record{}, unexpectedTokenTypeError{
			expected: []tokenType{tokenTypeNewline, tokenTypeEOF},
			actual:   p.curToken.tt,
		}
	}

	if err := p.bump(); err != nil {
		return record.Record{}, err
	}

	return record.Record{
		Name:  domain,
		Type:  *ty,
		Class: *class,
		TTL:   *ttl,
		Data:  data,
	}, nil
}

func (p *parser) parseOrigin() error {
	domain, err := p.parseDomain()
	if err != nil {
		return err
	}

	if !domain.IsAbsolute() {
		return errRelativeDomainOrigin
	}

	p.origin = domain

	return nil
}

func canonicalizeString(s string) string {
	runes := []rune(s)
	var buf strings.Builder

	for i := 0; i < len(runes); {
		if runes[i] == '\\' {
			i++

			if runes[i] >= '0' && runes[i] <= '9' {
				octet, err := strconv.ParseUint(string(runes[i:i+3]), 10, 8)
				if err != nil {
					panic(err)
				}

				_ = buf.WriteByte(byte(octet))
				i += 3
			}

			continue
		}

		_, _ = buf.WriteRune(runes[i])
		i++
	}

	return buf.String()
}
