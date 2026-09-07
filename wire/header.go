package wire

import (
	"encoding/binary"

	"github.com/MilkeeyCat/bdns/message"
)

const HeaderSize = 12

// https://www.rfc-editor.org/info/rfc1035/#section-4.1.1
type Header struct {
	// An identifier assigned by the program that generates any kind of query.
	// This identifier is copied the corresponding reply and can be used by the
	// requester to match up replies to outstanding queries.
	ID uint16
	// Specifies whether this message is a query(false), or a response(true).
	QR bool
	// This value is set by the originator of a query and copied into the
	// response.
	Opcode message.QueryOpcode
	// Authoritative Answer - the value is valid in responses, and specifies
	// that the responding name server is an authority for the domain name in
	// question section.
	//
	// Note that the contents of the answer section may have multiple owner
	// names because of aliases. The AA field corresponds to the name which
	// matches the query name, or the first owner name in the answer section.
	AA bool
	// TrunCation - specifies that this message was truncated due to length
	// greater than that permitted on the transmission channel.
	TC bool
	// Recursion Desired - this field may be set in a query and is copied into
	// the response. If RD field is set, it directs the name server to pursue
	// the query recursively.
	//
	// Recursive query support is optional.
	RD bool
	// Recursion Available - this field is set or cleared in a response, and
	// denotes whether recursive query support is available in the name server.
	RA    bool
	RCode message.ResponseCode
	// Number of entries in the question section.
	QDCount uint16
	// Number of resource records in the answer section.
	ANCount uint16
	// Number of name server resource records in the authority records section.
	NSCount uint16
	// Number of resource records in the additional records section.
	ARCount uint16
}

func ParseHeader(buf [HeaderSize]byte) (Header, error) {
	var opcode message.QueryOpcode

	switch (buf[2] & 0b0111_1000) >> 3 {
	case 0:
		opcode = message.QueryOpcodeStandard
	case 1:
		opcode = message.QueryOpcodeInverse
	case 2:
		opcode = message.QueryOpcodeServerStatus
	default:
		return Header{}, ErrInvalidMessage
	}

	var rcode message.ResponseCode

	switch buf[3] & 0b0000_1111 {
	case 0:
		rcode = message.ResponseCodeNone
	case 1:
		rcode = message.ResponseCodeFormatError
	case 2:
		rcode = message.ResponseCodeServerFailure
	case 3:
		rcode = message.ResponseCodeNameError
	case 4:
		rcode = message.ResponseCodeNotImplemented
	case 5:
		rcode = message.ResponseCodeRefused
	default:
		return Header{}, ErrInvalidMessage
	}

	return Header{
		ID:      binary.BigEndian.Uint16(buf[0:]),
		QR:      (buf[2] & 0b1000_0000) != 0,
		Opcode:  opcode,
		AA:      (buf[2] & 0b0000_0100) != 0,
		TC:      (buf[2] & 0b0000_0010) != 0,
		RD:      (buf[2] & 0b0000_0001) != 0,
		RA:      (buf[3] & 0b1000_0000) != 0,
		RCode:   rcode,
		QDCount: binary.BigEndian.Uint16(buf[4:]),
		ANCount: binary.BigEndian.Uint16(buf[6:]),
		NSCount: binary.BigEndian.Uint16(buf[8:]),
		ARCount: binary.BigEndian.Uint16(buf[10:]),
	}, nil
}
