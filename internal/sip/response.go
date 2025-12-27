package sip

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func BuildResponse(
	code int,
	reason string,
	req *Message,
) string {
	return fmt.Sprintf(
		"SIP/2.0 %d %s\r\n"+
			"Via: %s\r\n"+
			"From: %s\r\n"+
			"To: %s\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: %s\r\n"+
			"Content-Length: 0\r\n\r\n",
		code,
		reason,
		req.Headers["Via"],
		req.Headers["From"],
		addToTag(req.Headers["To"]),
		req.Headers["Call-ID"],
		req.Headers["CSeq"],
	)
}

func addToTag(to string) string {
	if strings.Contains(to, "tag=") {
		return to
	}
	return to + ";tag=server123" // todo: needs to be dynamic
}

func generateBranch() string {
	return "z9hG4bKproxy-" + uuid.New().String()
}

func buildProxyVia(proxyIp string, proxyPort int) string {
	return fmt.Sprintf(
		"SIP/2.0/UDP %s:%d;branch=%s",
		proxyIp,
		proxyPort,
		generateBranch(),
	)
}
