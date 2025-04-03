package main

import (
	"crypto"
	"testing"
)

func TestImportExport(t *testing.T) {
	tests := []struct {
		uri    string
		name   string
		secret string
		hash   crypto.Hash
		digits uint8
		period uint16
	}{
		{
			// test case from Google
			"otpauth://totp/ACME%20Co:john.doe@email.com?secret=HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ&issuer=ACME%20Co&algorithm=SHA1&digits=6&period=30",
			"ACME Co:john.doe@email.com",
			"HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ",
			crypto.SHA1,
			6,
			30,
		},
	}

	for _, test := range tests {
		e, err := EntryFromUri(test.uri)
		if err != nil {
			t.Errorf("parse uri failed: %v", err)
		} else {
			if test.name != e.Name {
				t.Errorf("expected name '%v' got '%v'", test.name, e.Name)
			}
			if test.secret != e.Secret {
				t.Errorf("expected secret '%v' got '%v'", test.secret, e.Secret)
			}
			if test.digits != e.Digits {
				t.Errorf("expected digits '%v' got '%v'", test.digits, e.Digits)
			}
			if test.period != e.Period {
				t.Errorf("expected period '%v' got '%v'", test.period, e.Period)
			}

			// encode it and reload it, then compare differences
			newstr, _ := EntryToUri(e)
			e2, err := EntryFromUri(newstr)
			if err != nil {
				t.Errorf("parse recoded uri failed: %v", err)
			} else {
				if e.Name != e2.Name {
					t.Errorf("expected recoded name '%v' got '%v'", e.Name, e2.Name)
				}
				if e.Secret != e2.Secret {
					t.Errorf("expected recoded secret '%v' got '%v'", e.Secret, e2.Secret)
				}
				if e.Period != e2.Period {
					t.Errorf("expected recoded period '%v' got '%v'", e.Period, e2.Period)
				}
				if e.Digits != e2.Digits {
					t.Errorf("expected recoded digits '%v' got '%v'", e.Digits, e2.Digits)
				}

			}

		}
	}

}
