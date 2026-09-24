package masterfile

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/record"
)

func TestParser(t *testing.T) {
	t.Run("resource record", testParseResourceRecord)
}

func testParseResourceRecord(t *testing.T) {
	tests := []struct {
		name  string
		input string
		rr    *record.Record
		err   error
	}{
		{
			name:  "CNAME type",
			input: "b.dummy.local.    12345    IN    CNAME    a.dummy.local.",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("b.dummy.local."),
				Type:  record.TypeCNAME,
				Class: record.ClassIN,
				TTL:   12345,
				Data: record.CNAMEData{
					CNAME: domain.NewDomainFromString("a.dummy.local."),
				},
			},
			err: nil,
		},
		{
			name:  "HINFO type",
			input: `a.dummy.local.    12345    IN    HINFO    "Intel Pentium"    "Goober OS"`,
			rr: &record.Record{
				Name:  domain.NewDomainFromString("a.dummy.local."),
				Type:  record.TypeHINFO,
				Class: record.ClassIN,
				TTL:   12345,
				Data: record.HINFOData{
					CPU: "Intel Pentium",
					OS:  "Goober OS",
				},
			},
			err: nil,
		},
		{
			name:  "MB type",
			input: "c.dummy.local.    86400    IN    MB    a.dummy.local.",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("c.dummy.local."),
				Type:  record.TypeMB,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.MBData{
					MADName: domain.NewDomainFromString("a.dummy.local."),
				},
			},
			err: nil,
		},
		{
			name:  "MG type",
			input: "c.dummy.local.    86400    IN    MG    a.dummy.local.",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("c.dummy.local."),
				Type:  record.TypeMG,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.MGData{
					MGMName: domain.NewDomainFromString("a.dummy.local."),
				},
			},
			err: nil,
		},
		{
			name:  "MINFO type",
			input: "c.dummy.local.    86400    IN    MINFO    a.dummy.local.    a.dummy.local.",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("c.dummy.local."),
				Type:  record.TypeMINFO,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.MINFOData{
					RMailbx: domain.NewDomainFromString("a.dummy.local."),
					EMailbx: domain.NewDomainFromString("a.dummy.local."),
				},
			},
			err: nil,
		},
		{
			name:  "MR type",
			input: "c.dummy.local.    86400    IN    MR    a.dummy.local.",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("c.dummy.local."),
				Type:  record.TypeMR,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.MRData{
					NewName: domain.NewDomainFromString("a.dummy.local."),
				},
			},
			err: nil,
		},
		{
			name:  "MX type",
			input: "c.dummy.local.    86400    IN    MX    69    a.dummy.local.",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("c.dummy.local."),
				Type:  record.TypeMX,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.MXData{
					Preference: 69,
					Exchange:   domain.NewDomainFromString("a.dummy.local."),
				},
			},
			err: nil,
		},
		{
			name:  "NULL type",
			input: `a.dummy.local.    86400    IN    NULL    \001\002\003\004\005`,
			rr: &record.Record{
				Name:  domain.NewDomainFromString("a.dummy.local."),
				Type:  record.TypeNULL,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.NULLData{
					Data: []byte{1, 2, 3, 4, 5},
				},
			},
			err: nil,
		},
		{
			name:  "NS type",
			input: "dummy.local.    86400    IN    NS    ns1.local.",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("dummy.local."),
				Type:  record.TypeNS,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.NSData{
					NSDName: domain.NewDomainFromString("ns1.local."),
				},
			},
			err: nil,
		},
		{
			name:  "PTR type",
			input: "c.dummy.local.    86400    IN    PTR    a.dummy.local.",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("c.dummy.local."),
				Type:  record.TypePTR,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.PTRData{
					PTRDName: domain.NewDomainFromString("a.dummy.local."),
				},
			},
			err: nil,
		},
		{
			name:  "SOA type",
			input: "dummy.local.    86400    IN    SOA    ns1.local. root.local. 2026091900 3600 1800 604800 86400",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("dummy.local."),
				Type:  record.TypeSOA,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.SOAData{
					MName:   domain.NewDomainFromString("ns1.local."),
					RName:   domain.NewDomainFromString("root.local."),
					Serial:  0x78c3b57c,
					Refresh: 3600,
					Retry:   1800,
					Expire:  604800,
					Minimum: 86400,
				},
			},
			err: nil,
		},
		{
			name:  "TXT type",
			input: "a.dummy.local.    86400    IN    TXT    str1 str2 str3",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("a.dummy.local."),
				Type:  record.TypeTXT,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.TXTData{
					Data: []string{"str1", "str2", "str3"},
				},
			},
			err: nil,
		},
		{
			name:  "A type",
			input: "a.dummy.local.    86400    IN    A    5.6.7.8",
			rr: &record.Record{
				Name:  domain.NewDomainFromString("a.dummy.local."),
				Type:  record.TypeA,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.AData{
					Address: netip.AddrFrom4([4]byte{5, 6, 7, 8}),
				},
			},
			err: nil,
		},
		{
			name:  "WKS type",
			input: `a.dummy.local.    86400    IN    WKS    9.10.11.12    6    \000\000\005\000\000\000\000\000\000\000\128`,
			rr: &record.Record{
				Name:  domain.NewDomainFromString("a.dummy.local."),
				Type:  record.TypeWKS,
				Class: record.ClassIN,
				TTL:   86400,
				Data: record.WKSData{
					Address:  netip.AddrFrom4([4]byte{9, 10, 11, 12}),
					Protocol: 6,
					Bitmap:   []byte{0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
				},
			},
			err: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := newParser(strings.NewReader(tc.input))
			require.NoError(t, err)

			rr, err := p.parseRR()

			assert.Equal(t, tc.err, err)

			if tc.rr != nil {
				assert.Equal(t, *tc.rr, rr)
			}
		})
	}
}
