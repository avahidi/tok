// Implementation of TOTP Key Uri Format
// see https://github.com/google/google-authenticator/wiki/Key-Uri-Format

package main

import (
	"fmt"
	"net/url"
	"strconv"
)

// EntryFromUri extracts an entry from a otpauth string
func EntryFromUri(uristr string) (*Entry, error) {
	uri, err := url.Parse(uristr)
	if err != nil {
		return nil, err
	}
	if !(uri.Scheme == "otpauth" || uri.Scheme == "apple-otpauth") || uri.Host != "totp" {
		return nil, fmt.Errorf("expected otpauth://totp/...")
	}

	query := uri.Query()
	digits, err := strconv.ParseInt(query.Get("digits"), 10, 8)
	if err != nil {
		return nil, err
	}

	period, err := strconv.ParseInt(query.Get("period"), 10, 8)
	if err != nil {
		return nil, err
	}

	return NewEntry(uri.Path[1:], query.Get("secret"), query.Get("algorithm"), "", int(period), int(digits))
}

// EntryToUri generates the otpauth string for this entry
func EntryToUri(e *Entry) (string, error) {
	return fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=issuer&algorithm=%s&digits=%d&period=%d",
		url.PathEscape(e.Name), e.Secret,
		hashToName(e.Hash), e.Digits, e.Period,
	), nil
}
