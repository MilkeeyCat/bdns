package wire

import (
	"encoding/binary"
	"fmt"

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

func EncodeHeader(header Header) [HeaderSize]byte {
	var buf [HeaderSize]byte
	var opcode uint8

	switch header.Opcode {
	case message.QueryOpcodeStandard:
		opcode = 0
	case message.QueryOpcodeInverse:
		opcode = 1
	case message.QueryOpcodeServerStatus:
		opcode = 2
	default:
		panic(fmt.Sprintf("unexpected message.QueryOpcode: %d", header.Opcode))
	}

	var rcode uint8

	switch header.RCode {
	case message.ResponseCodeNone:
		rcode = 0
	case message.ResponseCodeFormatError:
		rcode = 1
	case message.ResponseCodeServerFailure:
		rcode = 2
	case message.ResponseCodeNameError:
		rcode = 3
	case message.ResponseCodeNotImplemented:
		rcode = 4
	case message.ResponseCodeRefused:
		rcode = 5
	default:
		panic(fmt.Sprintf("unexpected message.ResponseCode: %d", header.RCode))
	}

	binary.BigEndian.PutUint16(buf[0:], header.ID)
	buf[2] |= boolToUint8(header.QR) << 7
	buf[2] |= opcode << 3
	buf[2] |= boolToUint8(header.AA) << 2
	buf[2] |= boolToUint8(header.TC) << 1
	buf[2] |= boolToUint8(header.RD)
	buf[3] |= boolToUint8(header.RA) << 7
	buf[3] |= rcode
	binary.BigEndian.PutUint16(buf[4:], header.QDCount)
	binary.BigEndian.PutUint16(buf[6:], header.ANCount)
	binary.BigEndian.PutUint16(buf[8:], header.NSCount)
	binary.BigEndian.PutUint16(buf[10:], header.ARCount)

	return buf
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}

	return 0
}
