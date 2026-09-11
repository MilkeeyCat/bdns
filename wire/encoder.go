package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/message"
	"github.com/MilkeeyCat/bdns/record"
)

var ErrInvalidDomain = errors.New("invalid domain")

type Encoder struct {
	buf          bytes.Buffer
	compressor   domainCompressor
	messageLimit uint16
}

type EncoderOption func(*Encoder)

func WithMessageLimit(limit uint16) EncoderOption {
	return func(e *Encoder) {
		e.messageLimit = limit
	}
}

func NewEncoder(options ...EncoderOption) *Encoder {
	e := new(Encoder)

	for _, opt := range options {
		opt(e)
	}

	return e
}

func (e *Encoder) Bytes() []byte {
	return e.buf.Bytes()
}

func (e *Encoder) EncodeDomain(domain domain.Domain) error {
	subdomain, offset, compressed, err := e.compressor.tryCompress(domain, uint16(e.buf.Len()))
	if err != nil {
		return err
	}

	if compressed {
		e.encodeDomain(subdomain)

		buf := e.buf.AvailableBuffer()

		buf = binary.BigEndian.AppendUint16(buf, 0b1100_0000_0000_0000|offset)
		_, _ = e.buf.Write(buf)

		return nil
	}

	e.encodeDomain(domain)

	return nil
}

func (e *Encoder) encodeDomain(domain domain.Domain) {
	for _, label := range domain {
		_ = e.buf.WriteByte(byte(len(label)))
		_, _ = e.buf.Write(label)
	}
}

func (e *Encoder) EncodeQuestion(question message.Question) error {
	if err := e.EncodeDomain(question.Name); err != nil {
		return err
	}

	buf := e.buf.AvailableBuffer()

	buf = binary.BigEndian.AppendUint16(buf, encodeQueryType(question.Type))
	buf = binary.BigEndian.AppendUint16(buf, encodeQueryClass(question.Class))
	_, _ = e.buf.Write(buf)

	return nil
}

func (e *Encoder) EncodeResourceRecord(rr record.Record) error {
	if err := e.EncodeDomain(rr.Name); err != nil {
		return err
	}

	buf := e.buf.AvailableBuffer()

	buf = binary.BigEndian.AppendUint16(buf, encodeType(rr.Type))
	buf = binary.BigEndian.AppendUint16(buf, encodeClass(rr.Class))
	buf = binary.BigEndian.AppendUint32(buf, rr.TTL)
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(rr.Data)))
	_, _ = e.buf.Write(buf)
	_, _ = e.buf.Write(rr.Data)

	return nil
}

func (e *Encoder) EncodeMessage(msg message.Message) error {
	headerPos := e.buf.Len()

	_, _ = e.buf.Write(make([]byte, HeaderSize))

	var (
		qdCount, anCount, nsCount, arCount uint16

		truncated bool
		err       error
	)

	qdCount, truncated, err = e.encodeQuestions(msg.Question)
	if err != nil || truncated {
		goto end
	}

	anCount, truncated, err = e.encodeRRs(msg.Answer)
	if err != nil || truncated {
		goto end
	}

	nsCount, truncated, err = e.encodeRRs(msg.Authority)
	if err != nil || truncated {
		goto end
	}

	arCount, truncated, err = e.encodeRRs(msg.Additional)

end:
	if err != nil {
		return err
	}

	msg.Header.TC = msg.Header.TC || truncated

	header := EncodeHeader(Header{
		Header:  msg.Header,
		QDCount: qdCount,
		ANCount: anCount,
		NSCount: nsCount,
		ARCount: arCount,
	})

	_ = copy(e.buf.Bytes()[headerPos:], header[:])

	return nil
}

func (e *Encoder) checkLimit(fn func() error) (bool, error) {
	if e.messageLimit > 0 && e.buf.Len() > int(e.messageLimit) {
		panic("message was bigger than the limit before checkLimit call")
	}

	beforeLen := e.buf.Len()

	if err := fn(); err != nil {
		return false, err
	}

	if e.messageLimit > 0 && e.buf.Len() > int(e.messageLimit) {
		e.buf.Truncate(beforeLen)

		return true, nil
	}

	return false, nil
}

func (e *Encoder) encodeQuestions(questions []message.Question) (uint16, bool, error) {
	for i, q := range questions {
		truncated, err := e.checkLimit(func() error {
			return e.EncodeQuestion(q)
		})
		if err != nil {
			return 0, false, err
		}

		if truncated {
			return uint16(i), true, nil
		}
	}

	return 0, false, nil
}

func (e *Encoder) encodeRRs(rrs []record.Record) (uint16, bool, error) {
	for i, rr := range rrs {
		truncated, err := e.checkLimit(func() error {
			return e.EncodeResourceRecord(rr)
		})
		if err != nil {
			return 0, false, err
		}

		if truncated {
			return uint16(i), true, nil
		}
	}

	return 0, false, nil
}

type domainCompressor struct {
	root domainCompressorNode
}

func (c *domainCompressor) tryCompress(domain domain.Domain, offset uint16) (domain.Domain, uint16, bool, error) {
	if !domain.IsAbsolute() {
		return nil, 0, false, ErrInvalidDomain
	}

	domain = domain[:len(domain)-1]

	node := &c.root
	i := len(domain)

	for _, label := range slices.Backward(domain) {
		next := node.find(label)
		if next == nil {
			break
		}

		i -= 1
		node = next
	}

	domain = domain[:i]

	{
		offsets := make([]uint16, len(domain))

		for i, label := range domain {
			offsets[i] = offset
			offset += uint16(len(label)) + 1
		}

		cur := node

		for i, label := range slices.Backward(domain) {
			cur = cur.insert(label, offsets[i])
		}
	}

	if node == &c.root {
		return nil, 0, false, nil
	}

	return domain, node.offset, true, nil
}

type domainCompressorNode struct {
	label    domain.Label
	offset   uint16
	children []*domainCompressorNode
}

func (n *domainCompressorNode) find(label domain.Label) *domainCompressorNode {
	for _, child := range n.children {
		if child.label.Equals(label) {
			return child
		}
	}

	return nil
}

func (n *domainCompressorNode) insert(label domain.Label, offset uint16) *domainCompressorNode {
	for _, child := range n.children {
		if child.label.Equals(label) {
			panic("node is already inserted")
		}
	}

	child := &domainCompressorNode{
		label:    label,
		offset:   offset,
		children: nil,
	}

	n.children = append(n.children, child)

	return child
}

func encodeType(ty record.Type) uint16 {
	switch ty {
	case record.TypeA:
		return 1
	case record.TypeNS:
		return 2
	case record.TypeMD:
		return 3
	case record.TypeMF:
		return 4
	case record.TypeCNAME:
		return 5
	case record.TypeSOA:
		return 6
	case record.TypeMB:
		return 7
	case record.TypeMG:
		return 8
	case record.TypeMR:
		return 9
	case record.TypeNULL:
		return 10
	case record.TypeWKS:
		return 11
	case record.TypePTR:
		return 12
	case record.TypeHINFO:
		return 13
	case record.TypeMINFO:
		return 14
	case record.TypeMX:
		return 15
	case record.TypeTXT:
		return 16
	default:
		panic(fmt.Sprintf("unexpected record.Type: %d", ty))
	}
}

func encodeClass(class record.Class) uint16 {
	switch class {
	case record.ClassIN:
		return 1
	case record.ClassCS:
		return 2
	case record.ClassCH:
		return 3
	case record.ClassHS:
		return 4
	default:
		panic(fmt.Sprintf("unexpected record.Class: %d", class))
	}
}

func encodeQueryType(qty message.QueryType) uint16 {
	switch qty {
	case message.QueryTypeAXFR:
		return 252
	case message.QueryTypeMAILB:
		return 253
	case message.QueryTypeMAILA:
		return 254
	case message.QueryTypeAll:
		return 255
	default:
		return encodeType(record.Type(qty))
	}
}

func encodeQueryClass(qclass message.QueryClass) uint16 {
	switch qclass {
	case message.QueryClassAny:
		return 255
	default:
		return encodeClass(record.Class(qclass))
	}
}
