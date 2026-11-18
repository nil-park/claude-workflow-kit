package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"testing"
)

type capturedRequest struct {
	ID     int
	Method string
	Params map[string]any
}

func startFakeServer(t *testing.T, handler func(conn net.Conn, req capturedRequest)) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		for {
			payload, err := readFrame(r)
			if err != nil {
				return
			}
			var raw struct {
				ID     int            `json:"id"`
				Method string         `json:"method"`
				Params map[string]any `json:"params"`
			}
			if err := json.Unmarshal(payload, &raw); err != nil {
				return
			}
			handler(conn, capturedRequest{ID: raw.ID, Method: raw.Method, Params: raw.Params})
		}
	}()

	return ln.Addr().(*net.TCPAddr).Port
}

func writeFakeFrame(w io.Writer, msg any) {
	b, _ := json.Marshal(msg)
	fmt.Fprintf(w, "Content-Length: %d\r\n\r\n%s", len(b), b)
}

func writeResult(w io.Writer, id int, result any) {
	writeFakeFrame(w, map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
}

func writeError(w io.Writer, id, code int, message string) {
	writeFakeFrame(w, map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error":   map[string]any{"code": code, "message": message},
	})
}
