package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

const httpContentType = "application/json"

type Relay struct {
	Endpoint string

	HTTPClient http.Client
}

func (r *Relay) Relay(ctx context.Context, body io.Reader, w http.ResponseWriter) error {
	req, err := http.NewRequest(http.MethodPost, r.Endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", httpContentType)
	req.Header.Set("Accept", httpContentType)
	req = req.WithContext(ctx)

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

type RPCRequest struct {
	Method string `json:"method"`
}

type RPCHandler struct {
	// MaxContentLength is the request size limit (optional)
	MaxContentLength int64

	ShouldRelayMethod func(method string) bool
	Relay             func(ctx context.Context, body io.Reader, w http.ResponseWriter) error
}

func (h *RPCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Convert http.Error(...) output into actual JSONRPC error responses?
	if r.Method == http.MethodGet && r.ContentLength == 0 && r.URL.RawQuery == "" {
		// Ignore empty GET requests
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.MaxContentLength > 0 && r.ContentLength > h.MaxContentLength {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	var body io.Reader = r.Body
	if h.MaxContentLength > 0 {
		body = io.LimitReader(r.Body, h.MaxContentLength)
	}
	defer r.Body.Close()

	// Load request into memory to relay later
	var raw json.RawMessage
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		http.Error(w, "JSON parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse method
	var req RPCRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		http.Error(w, "method decode error: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !h.ShouldRelayMethod(req.Method) {
		// XXX: Implement buffering
		http.Error(w, "method relay rejected", http.StatusForbidden)
		return
	}

	if err := h.Relay(r.Context(), bytes.NewReader(raw), w); err != nil {
		http.Error(w, "failed to relay: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("content-type", httpContentType)

}
