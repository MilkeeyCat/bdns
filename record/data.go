package record

import (
	"net/netip"

	"github.com/MilkeeyCat/bdns/domain"
)

type Data interface {
	data()
}

type CNAMEData struct {
	// Canonical or primary name for the owner. The owner name is an alias.
	CNAME domain.Domain
}

func (d CNAMEData) data() {}

type HINFOData struct {
	CPU string
	OS  string
}

func (d HINFOData) data() {}

type MBData struct {
	// Host which has the specified mailbox.
	MADName domain.Domain
}

func (d MBData) data() {}

type MGData struct {
	// Mailbox which is a member of the mail group specified by the domain name.
	MGMName domain.Domain
}

func (d MGData) data() {}

type MINFOData struct {
	// Mailbox which is responsible for the mailing list or mailbox. If this
	// domain name names the root, the owner of the MINFO RR is responsible for
	// itself. Note that many existing mailing lists use a mailbox X-request
	// for the RMAILBX field of mailing list X, e.g., Msgroup-request for
	// Msgroup. This field provides a more general mechanism.
	RMailbx domain.Domain

	// Mailbox which is to receive error messages related to the mailing list or
	// mailbox specified by the owner of the MINFO RR (similar to the ERRORS-TO:
	// field which has been proposed). If this domain name names the root,
	// errors should be returned to the sender of the message.
	EMailbx domain.Domain
}

func (d MINFOData) data() {}

type MRData struct {
	// Mailbox which is the proper rename of the specified mailbox.
	NewName domain.Domain
}

func (d MRData) data() {}

type MXData struct {
	// Preference given to this RR among others at the same owner. Lower values
	// are preferred.
	Preference uint16

	// Host willing to act as a mail exchange for the owner name.
	Exchange domain.Domain
}

func (d MXData) data() {}

type NULLData struct {
	Data []byte
}

func (d NULLData) data() {}

type NSData struct {
	// Host which should be authoritative for the specified class and domain.
	NSDName domain.Domain
}

func (d NSData) data() {}

type PTRData struct {
	// Domain name which points to some location in the domain name space.
	PTRDName domain.Domain
}

func (d PTRData) data() {}

type SOAData struct {
	// Domain name of the name server that was the original or primary source of
	// data for this zone.
	MName domain.Domain

	// Domain name which specifies the mailbox of the person responsible for
	// this zone.
	RName domain.Domain

	// Version number of the original copy of the zone. Zone transfers preserve
	// this value. This value wraps and should be compared using sequence space
	// arithmetic.
	Serial uint32

	// Time interval before the zone should be refreshed.
	Refresh uint32

	// Time interval that should elapse before a failed refresh should be
	// retried.
	Retry uint32

	// Time value that specifies the upper limit on the time interval that can
	// elapse before the zone is no longer authoritative.
	Expire uint32

	// Minimum TTL field that should be exported with any RR from this zone.
	Minimum uint32
}

func (d SOAData) data() {}

type TXTData struct {
	Data []string
}

func (d TXTData) data() {}

type AData struct {
	Address netip.Addr
}

func (d AData) data() {}

type WKSData struct {
	Address netip.Addr

	// IP protocol number.
	Protocol uint8
	Bitmap   []byte
}

func (d WKSData) data() {}
