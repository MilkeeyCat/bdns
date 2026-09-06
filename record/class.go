package record

type Class uint8

const (
	// The Internet.
	ClassIN Class = iota

	// The CSNET class (Obsolete - used only for examples in some obsolete RFCs).
	ClassCS

	// The CHAOS class.
	ClassCH

	// Hesiod [Dyer 87].
	ClassHS
)
