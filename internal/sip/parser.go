package sip

import (
	"strconv"
	"strings"
)

type Message struct {
	Method     string
	StatusCode int
	Headers    map[string]string
	Body       string
	Raw        string
}

func ParseMessage(data []byte) *Message {
	raw := string(data)
	lines := strings.Split(raw, "\r\n")

	msg := &Message{
		Headers: make(map[string]string),
		Raw:     raw,
	}

	if strings.HasPrefix(lines[0], "SIP/2.0") {
		msg.Method = "RESPONSE"
		parts := strings.Split(lines[0], " ")
		if len(parts) >= 2 {
			msg.StatusCode, _ = strconv.Atoi(parts[1])
		}
	} else {
		parts := strings.Split(lines[0], " ")
		msg.Method = parts[0]
	}

	i := 1
	for ; i < len(lines); i++ {
		if lines[i] == "" {
			break
		}
		kv := strings.SplitN(lines[i], ":", 2)
		if len(kv) == 2 {
			msg.Headers[strings.TrimSpace(kv[0])] =
				strings.TrimSpace(kv[1])
		}
	}

	if i+1 < len(lines) {
		msg.Body = strings.Join(lines[i+1:], "\r\n")
	}

	return msg
}
