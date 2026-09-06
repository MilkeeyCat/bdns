package record

type Type uint8

const (
	// Host address.
	TypeA Type = iota

	// Authoritative name server.
	TypeNS

	// Mail destination (Obsolete - use [TypeMX]).
	TypeMD

	// Mail forwarder (Obsolete - use [TypeMX]).
	TypeMF

	// Canonical name for an alias.
	TypeCNAME

	// Marks the start of a zone of authority.
	TypeSOA

	// Mailbox domain name (EXPERIMENTAL).
	TypeMB

	// Mail group member (EXPERIMENTAL).
	TypeMG

	// Mail rename domain name (EXPERIMENTAL).
	TypeMR

	// null RR (EXPERIMENTAL).
	TypeNULL

	// Well known service description.
	TypeWKS

	// Domain name pointer.
	TypePTR

	// Host information.
	TypeHINFO

	// Mailbox or mail list information.
	TypeMINFO

	// Mail exchange.
	TypeMX

	// Text strings.
	TypeTXT
)
