package masterfile_test

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/masterfile"
	"github.com/MilkeeyCat/bdns/record"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		rrs   []record.Record
	}{
		{
			input: `
$ORIGIN LOCAL.

@   86400    IN  SOA     VENERA      Action\.domains (
								 20     ; SERIAL
								 7200   ; REFRESH
								 600    ; RETRY
								 3600000; EXPIRE
								 60)    ; MINIMUM

		     NS      A.ISI.EDU.
		     NS      VENERA
		     NS      VAXA
		     MX      10      VENERA
		     MX      20      VAXA

A            A       26.3.0.103

VENERA       A       10.1.0.52
		     A       128.9.0.32

VAXA         A       10.2.0.27
		     A       128.9.0.33
`,
			rrs: []record.Record{
				{
					Name:  domain.NewDomainFromString("LOCAL."),
					Type:  record.TypeSOA,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.SOAData{
						MName: domain.NewDomainFromString("VENERA.LOCAL."),
						RName: domain.Domain{
							{'A', 'c', 't', 'i', 'o', 'n', '.', 'd', 'o', 'm', 'a', 'i', 'n', 's'},
							{'L', 'O', 'C', 'A', 'L'},
							{},
						},
						Serial:  0x14,
						Refresh: 7200,
						Retry:   600,
						Expire:  3600000,
						Minimum: 60,
					},
				},
				{
					Name:  domain.NewDomainFromString("LOCAL."),
					Type:  record.TypeNS,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.NSData{
						NSDName: domain.NewDomainFromString("A.ISI.EDU."),
					},
				},
				{
					Name:  domain.NewDomainFromString("LOCAL."),
					Type:  record.TypeNS,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.NSData{
						NSDName: domain.NewDomainFromString("VENERA.LOCAL."),
					},
				},
				{
					Name:  domain.NewDomainFromString("LOCAL."),
					Type:  record.TypeNS,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.NSData{
						NSDName: domain.NewDomainFromString("VAXA.LOCAL."),
					},
				},
				{
					Name:  domain.NewDomainFromString("LOCAL."),
					Type:  record.TypeMX,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.MXData{
						Preference: 10,
						Exchange:   domain.NewDomainFromString("VENERA.LOCAL."),
					},
				},
				{
					Name:  domain.NewDomainFromString("LOCAL."),
					Type:  record.TypeMX,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.MXData{
						Preference: 20,
						Exchange:   domain.NewDomainFromString("VAXA.LOCAL."),
					},
				},
				{
					Name:  domain.NewDomainFromString("A.LOCAL."),
					Type:  record.TypeA,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.AData{
						Address: netip.AddrFrom4([4]byte{26, 3, 0, 103}),
					},
				},
				{
					Name:  domain.NewDomainFromString("VENERA.LOCAL."),
					Type:  record.TypeA,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.AData{
						Address: netip.AddrFrom4([4]byte{10, 1, 0, 52}),
					},
				},
				{
					Name:  domain.NewDomainFromString("VENERA.LOCAL."),
					Type:  record.TypeA,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.AData{
						Address: netip.AddrFrom4([4]byte{128, 9, 0, 32}),
					},
				},
				{
					Name:  domain.NewDomainFromString("VAXA.LOCAL."),
					Type:  record.TypeA,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.AData{
						Address: netip.AddrFrom4([4]byte{10, 2, 0, 27}),
					},
				},
				{
					Name:  domain.NewDomainFromString("VAXA.LOCAL."),
					Type:  record.TypeA,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.AData{
						Address: netip.AddrFrom4([4]byte{128, 9, 0, 33}),
					},
				},
			},
		},
		{
			input: `
$ORIGIN LOCAL.

MOE     86400 IN MB      A.ISI.EDU.
LARRY            MB      A.ISI.EDU.
CURLEY           MB      A.ISI.EDU.
STOOGES          MG      MOE
		         MG      LARRY
		         MG      CURLEY
`,
			rrs: []record.Record{
				{
					Name:  domain.NewDomainFromString("MOE.LOCAL."),
					Type:  record.TypeMB,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.MBData{
						MADName: domain.NewDomainFromString("A.ISI.EDU."),
					},
				},
				{
					Name:  domain.NewDomainFromString("LARRY.LOCAL."),
					Type:  record.TypeMB,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.MBData{
						MADName: domain.NewDomainFromString("A.ISI.EDU."),
					},
				},
				{
					Name:  domain.NewDomainFromString("CURLEY.LOCAL."),
					Type:  record.TypeMB,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.MBData{
						MADName: domain.NewDomainFromString("A.ISI.EDU."),
					},
				},
				{
					Name:  domain.NewDomainFromString("STOOGES.LOCAL."),
					Type:  record.TypeMG,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.MGData{
						MGMName: domain.NewDomainFromString("MOE.LOCAL."),
					},
				},
				{
					Name:  domain.NewDomainFromString("STOOGES.LOCAL."),
					Type:  record.TypeMG,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.MGData{
						MGMName: domain.NewDomainFromString("LARRY.LOCAL."),
					},
				},
				{
					Name:  domain.NewDomainFromString("STOOGES.LOCAL."),
					Type:  record.TypeMG,
					Class: record.ClassIN,
					TTL:   86400,
					Data: record.MGData{
						MGMName: domain.NewDomainFromString("CURLEY.LOCAL."),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		rrs, err := masterfile.Parse(strings.NewReader(tc.input))

		require.NoError(t, err)
		assert.Equal(t, tc.rrs, rrs)
	}
}
