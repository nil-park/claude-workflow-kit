package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const rpcTimeout = 10 * time.Second

func runSend(args []string, stdout, stderr io.Writer) error {
	flags := newFlagSet("send", stderr)
	port := flags.Int("port", 0, "TCP port of the Copilot UI server")
	from := flags.String("from", "", "sender session name")
	message := flags.String("message", "", "message body")
	file := flags.String("file", "", "path of a file holding the message body")
	handshake := flags.Bool("handshake", false, "prepend the reply handshake guide")
	if err := parseFlags(flags, args); err != nil {
		return err
	}
	given := givenFlags(flags)
	if !given["port"] {
		return errors.New("--port is required")
	}
	if *port < 1 || *port > 65535 {
		return fmt.Errorf("--port must be between 1 and 65535: %d", *port)
	}
	if !given["from"] {
		return errors.New("--from is required")
	}
	if err := checkFromName(*from); err != nil {
		return err
	}
	body, err := readBody(given, *message, *file, *handshake, *from, *port)
	if err != nil {
		return err
	}
	msgID, err := sendPrompt(*port, wrapBody(*from, body), rpcTimeout)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, msgID)
	return nil
}

// givenFlags uses flags.Visit so an explicit empty value such as --message ""
// counts as given instead of as missing.
func givenFlags(flags *flag.FlagSet) map[string]bool {
	given := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { given[f.Name] = true })
	return given
}

func checkFromName(name string) error {
	if name == "" {
		return errors.New("--from must not be empty")
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return fmt.Errorf("--from must contain only lowercase ASCII letters, digits, and '-': %q", name)
		}
	}
	return nil
}

func readBody(given map[string]bool, message, file string, handshake bool, from string, port int) (string, error) {
	if given["message"] && given["file"] {
		return "", errors.New("pass at most one of --message and --file")
	}
	hasMsg := given["message"] || given["file"]
	if !hasMsg && !handshake {
		return "", errors.New("pass --handshake, --message, or --file")
	}
	var parts []string
	if handshake {
		parts = append(parts, handshakeBody(from, port))
	}
	if hasMsg {
		body := message
		if given["file"] {
			raw, err := os.ReadFile(file)
			if err != nil {
				return "", err
			}
			if !utf8.Valid(raw) {
				return "", fmt.Errorf("%s is not valid UTF-8", file)
			}
			body = string(raw)
		}
		if strings.TrimSpace(body) == "" {
			return "", errors.New("message body is empty")
		}
		parts = append(parts, body)
	}
	return strings.Join(parts, "\n\n"), nil
}

// wrapBody escapes closing tags in the body so it cannot end the wrapper
// early. from needs no escaping because checkFromName allows only [a-z0-9-].
func wrapBody(from, body string) string {
	escaped := strings.ReplaceAll(body, "</cross-session-message", `<\/cross-session-message`)
	return fmt.Sprintf("<cross-session-message from=\"%s\">\n%s\n</cross-session-message>", from, escaped)
}
