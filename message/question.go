package message

import (
	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/record"
)

// QueryType is a superset of [record.Type]; every [record.Type] is a valid
// [QueryType].
type QueryType record.Type

const (
	// Request for a transfer of an entire zone.
	QueryTypeAXFR QueryType = iota + QueryType(record.TypeTXT) + 1

	// Request for mailbox-related records ([record.TypeMB], [record.TypeMG] or [record.TypeMR]).
	QueryTypeMAILB

	// Request for mail agent RRs (Obsolete - see [record.TypeMX]).
	QueryTypeMAILA

	// Request for all records.
	QueryTypeAll
)

// QueryClass is a superset of [record.Class]; every [record.Class] is a valid
// [QueryClass].
type QueryClass record.Class

const (
	// Any class.
	QueryClassAny QueryClass = iota + QueryClass(record.ClassHS) + 1
)

// https://www.rfc-editor.org/info/rfc1035/#section-4.1.2
type Question struct {
	Name  domain.Domain
	Type  QueryType
	Class QueryClass
}
