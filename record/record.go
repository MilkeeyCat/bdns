package record

import "github.com/MilkeeyCat/bdns/domain"

type Record struct {
	Name  domain.Domain
	Type  Type
	Class Class
	TTL   uint32
	Data  []byte
}
