package domain

import (
	"bytes"
	"strings"
)

type Label []byte

func (l Label) Equals(other Label) bool {
	if len(l) != len(other) {
		return false
	}

	for i := range l {
		if lowerASCII(l[i]) != lowerASCII(other[i]) {
			return false
		}
	}

	return true
}

func lowerASCII(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}

	return b
}

func (l Label) IsRoot() bool {
	return len(l) == 0
}

type Domain []Label

func NewDomainFromString(s string) Domain {
	parts := bytes.Split([]byte(s), []byte{'.'})
	labels := make([]Label, len(parts))

	for i, part := range parts {
		labels[i] = part
	}

	return labels
}

func (d Domain) IsAbsolute() bool {
	if len(d) > 0 {
		return d[len(d)-1].IsRoot()
	}

	return false
}

func (d Domain) String() string {
	if len(d) == 1 && len(d[0]) == 0 {
		return "."
	}

	var buf strings.Builder

	for i, label := range d {
		_, _ = buf.WriteString(string(label))

		if len(d)-1 > i {
			buf.WriteByte('.')
		}
	}

	return buf.String()
}
