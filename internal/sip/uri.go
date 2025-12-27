package sip

import "strings"

func ExtractUser(uri string) string {
	uri = strings.TrimSpace(uri)
	uri = strings.Trim(uri, "<>")
	if strings.HasPrefix(uri, "sip:") {
		uri = strings.TrimPrefix(uri, "sip:")
	}
	parts := strings.Split(uri, "@")
	return parts[0]
}
