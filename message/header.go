package message

type QueryOpcode uint8

const (
	// A standard query (QUERY).
	QueryOpcodeStandard QueryOpcode = iota

	// An inverse query (IQUERY).
	QueryOpcodeInverse

	// A server status request (STATUS).
	QueryOpcodeServerStatus
)

type ResponseCode uint8

const (
	// No error condition.
	ResponseCodeNone ResponseCode = iota

	// The name server was unable to interpret the query.
	ResponseCodeFormatError

	// The name server was unable to process this query due to a problem with
	// the name server.
	ResponseCodeServerFailure

	// Meaningful only for responses from an authoritative name server, this
	// code signifies that the domain name referenced in the query does not
	// exist.
	ResponseCodeNameError

	// The name server does not support the requested kind of query.
	ResponseCodeNotImplemented

	// The name server refuses to perform the specified operation for policy
	// reasons. For example, a name server may not wish to provide the
	// information to the particular requester, or a name server may not wish to
	// perform a particular operation(e.g., zone transfer) for particular data.
	ResponseCodeRefused
)

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
	Opcode QueryOpcode
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
	RCode ResponseCode
}
