package wire

import "github.com/MilkeeyCat/bdns/record"

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
