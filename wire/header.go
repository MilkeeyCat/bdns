package wire

import (
	"encoding/binary"

	"github.com/MilkeeyCat/bdns/message"
)

const HeaderSize = 12

type Header struct {
	message.Header

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
