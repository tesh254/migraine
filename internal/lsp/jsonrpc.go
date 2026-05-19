package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type jsonrpcHandler struct {
	server *Server
}

type jsonrpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func RunStdioServer(server *Server) error {
	handler := &jsonrpcHandler{server: server}
	reader := bufio.NewReaderSize(os.Stdin, 8192)
	writer := os.Stdout

	for {
		contentLength, err := readHeader(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			continue
		}

		body := make([]byte, contentLength)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			continue
		}

		var msg jsonrpcMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			continue
		}

		go func(msg jsonrpcMessage) {
			response := handler.handleMessage(context.Background(), msg)
			if response != nil {
				writeResponse(writer, response)
			}
		}(msg)
	}
}

func readHeader(reader *bufio.Reader) (int, error) {
	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				n, err := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err == nil {
					contentLength = n
				}
			}
		}
	}
	if contentLength == 0 {
		return 0, fmt.Errorf("missing Content-Length header")
	}
	return contentLength, nil
}

func writeResponse(writer io.Writer, msg *jsonrpcMessage) {
	body, err := json.Marshal(msg)
	if err != nil {
		return
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	writer.Write([]byte(header))
	writer.Write(body)
}

func (h *jsonrpcHandler) handleMessage(ctx context.Context, msg jsonrpcMessage) *jsonrpcMessage {
	if msg.Method == "" {
		return nil
	}

	result, err := h.server.Handle(ctx, msg.Method, msg.Params)

	if msg.ID == nil {
		return nil
	}

	resp := &jsonrpcMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
	}

	if err != nil {
		resp.Error = &jsonrpcError{
			Code:    -32603,
			Message: err.Error(),
		}
	} else {
		resp.Result = result
	}

	return resp
}