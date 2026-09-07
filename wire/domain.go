package wire

import "github.com/MilkeeyCat/bdns/domain"

func ParseDomain(buf []byte, offset uint) (domain.Domain, uint8, error) {
	if len(buf) < int(offset) {
		return nil, 0, ErrShortMessage
	}

	return parseDomain(buf, buf[offset:], nil, 0, make(map[uint16]struct{}))
}

func parseDomain(buf, cur []byte, domain domain.Domain, size uint8, offsets map[uint16]struct{}) (domain.Domain, uint8, error) {
	for {
		if len(cur) < 1 {
			return nil, 0, ErrShortMessage
		}

		length := cur[0]

		switch length & 0b1100_0000 {
		case 0b1100_0000:
			if len(cur) < 2 {
				return nil, 0, ErrShortMessage
			}

			offset := (uint16(length&0b0011_1111) << 8) | uint16(cur[1])

			if _, ok := offsets[offset]; ok {
				return nil, 0, ErrInvalidMessage
			}

			if len(buf) <= int(offset) {
				return nil, 0, ErrInvalidMessage
			}

			offsets[offset] = struct{}{}

			domain, _, err := parseDomain(buf, buf[offset:], domain, size, offsets)
			if err != nil {
				return nil, 0, err
			}

			return domain, size + 2, err

		case 0:
			if len(cur) < int(length) {
				return nil, 0, ErrShortMessage
			}

			domain = append(domain, cur[1:length+1])
			cur = cur[length+1:]
			size += length + 1

			if length == 0 {
				return domain, size, nil
			}

		default:
			return nil, 0, ErrInvalidMessage
		}
	}
}
