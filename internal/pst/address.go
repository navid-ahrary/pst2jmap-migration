package pst

import (
	"net/mail"
)

func ParseAddress(
	raw string,
) string {

	addr, err :=
		mail.ParseAddress(raw)

	if err != nil {
		return raw
	}

	return addr.Address
}
