package wire

import "github.com/MilkeeyCat/bdns/message"

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

func ParseQuestion(buf []byte, offset uint) (message.Question, uint, error) {
	domain, size, err := ParseDomain(buf, offset)
	if err != nil {
		return message.Question{}, 0, err
	}

	buf = buf[offset+uint(size):]

	if len(buf) < 4 {
		return message.Question{}, 0, ErrShortMessage
	}

	queryType, err := parseQueryType((uint16(buf[0]) << 8) | uint16(buf[1]))
	if err != nil {
		return message.Question{}, 0, err
	}

	queryClass, err := parseQueryClass((uint16(buf[2]) << 8) | uint16(buf[3]))
	if err != nil {
		return message.Question{}, 0, err
	}

	return message.Question{
		Name:  domain,
		Type:  queryType,
		Class: queryClass,
	}, uint(size) + 4, nil
}
