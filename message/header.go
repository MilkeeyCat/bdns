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
