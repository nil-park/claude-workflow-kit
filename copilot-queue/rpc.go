package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	expectedProtocolVersion = 3
	maxFrameBytes           = 4 << 20
)

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type rpcResponse struct {
	ID     *int            `json:"id"`
	Method string          `json:"method"`
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// sendPrompt connects to the Copilot UI server, checks protocolVersion and the
// foreground session, and delivers prompt without retrying because
// session.send is not idempotent.
func sendPrompt(port int, prompt, displayPrompt string, timeout time.Duration) (string, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		if isConnRefused(err) {
			return "", fmt.Errorf("no Copilot UI server is listening on port %d: %w", port, err)
		}
		return "", fmt.Errorf("connecting to Copilot UI server on port %d: %w", port, err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	var pingRes struct {
		ProtocolVersion int `json:"protocolVersion"`
	}
	if err := callRPC(conn, reader, 1, "ping", map[string]string{"message": "ping"}, &pingRes, timeout); err != nil {
		return "", err
	}
	if pingRes.ProtocolVersion != expectedProtocolVersion {
		return "", fmt.Errorf("unexpected Copilot protocolVersion %d (expected %d)", pingRes.ProtocolVersion, expectedProtocolVersion)
	}

	var fgRes struct {
		SessionID string `json:"sessionId"`
	}
	if err := callRPC(conn, reader, 2, "session.getForeground", struct{}{}, &fgRes, timeout); err != nil {
		return "", err
	}
	if fgRes.SessionID == "" {
		return "", errors.New("no foreground session")
	}

	sendParams := struct {
		SessionID     string `json:"sessionId"`
		Prompt        string `json:"prompt"`
		DisplayPrompt string `json:"displayPrompt"`
	}{
		SessionID:     fgRes.SessionID,
		Prompt:        prompt,
		DisplayPrompt: displayPrompt,
	}
	var sendRes struct {
		MessageID string `json:"messageId"`
	}
	if err := callRPC(conn, reader, 3, "session.send", sendParams, &sendRes, timeout); err != nil {
		return "", err
	}
	return sendRes.MessageID, nil
}

func isConnRefused(err error) bool {
	// 10061 is WSAECONNREFUSED on Windows.
	return errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.Errno(10061))
}

func callRPC(conn net.Conn, reader *bufio.Reader, id int, method string, params, target any, timeout time.Duration) error {
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	if err := writeFrame(conn, rpcRequest{JSONRPC: "2.0", ID: id, Method: method, Params: params}); err != nil {
		return fmt.Errorf("sending %s: %w", method, err)
	}
	for {
		payload, err := readFrame(reader)
		if err != nil {
			return fmt.Errorf("reading %s response: %w", method, err)
		}
		var resp rpcResponse
		if err := json.Unmarshal(payload, &resp); err != nil {
			return fmt.Errorf("decoding %s response: %w", method, err)
		}
		// Skip server-initiated requests and notifications even if their id collides.
		if resp.Method != "" || resp.ID == nil || *resp.ID != id {
			continue
		}
		if resp.Error != nil {
			return fmt.Errorf("%s failed (%d): %s", method, resp.Error.Code, resp.Error.Message)
		}
		if err := json.Unmarshal(resp.Result, target); err != nil {
			return fmt.Errorf("decoding %s result: %w", method, err)
		}
		return nil
	}
}

func writeFrame(w io.Writer, req rpcRequest) error {
	var body bytes.Buffer
	enc := json.NewEncoder(&body)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(req); err != nil {
		return err
	}
	payload := bytes.TrimSuffix(body.Bytes(), []byte("\n"))
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(payload))
	if _, err := io.WriteString(w, header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func readFrame(r *bufio.Reader) ([]byte, error) {
	contentLen := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(k), "Content-Length") {
			n, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil || n < 0 {
				return nil, fmt.Errorf("invalid Content-Length %q", v)
			}
			if n > maxFrameBytes {
				return nil, fmt.Errorf("Content-Length %d exceeds limit %d", n, maxFrameBytes)
			}
			contentLen = n
		}
	}
	if contentLen < 0 {
		return nil, errors.New("missing Content-Length header")
	}
	payload := make([]byte, contentLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}
