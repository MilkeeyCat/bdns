package masterfile

import (
	"errors"
	"io"

	"github.com/MilkeeyCat/bdns/record"
)

func Parse(rd io.Reader) ([]record.Record, error) {
	p, err := newParser(rd)
	if err != nil {
		return nil, err
	}

	var rrs []record.Record

	for p.curToken.tt != tokenTypeEOF {
		if p.curToken.tt == tokenTypeNewline {
			if err := p.bump(); err != nil {
				return nil, err
			}

			continue
		}

		if p.curToken.tt == tokenTypeIndent && p.peekToken.tt == tokenTypeNewline {
			for range 2 {
				if err := p.bump(); err != nil {
					return nil, err
				}
			}

			continue
		}

		str, err := p.currentStringRaw()
		if err == nil {
			switch str {
			case "$ORIGIN":
				if err := p.bump(); err != nil {
					return nil, err
				}

				if err := p.parseOrigin(); err != nil {
					return nil, err
				}

			case "$INCLUDE":
				// TODO: add support $INCLUDE
				return nil, errors.New("$INCLUDE is not supported(yet)")

			default:
				goto parse_rr
			}

			continue
		}

	parse_rr:
		rr, err := p.parseRR()
		if err != nil {
			return nil, err
		}

		rrs = append(rrs, rr)
	}

	return rrs, nil
}
