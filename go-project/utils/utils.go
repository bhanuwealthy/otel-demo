package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type errResponse struct {
	Message string `json:"message"`
}

func SendRequest(ctx context.Context, method string, url string, data []byte) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	client := http.Client{
		// Wrap the Transport with one that starts a span and injects the span context
		// into the outbound request headers.
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	return client.Do(request)
}

func WriteResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("encode response error: %v", err)
	}
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	WriteResponse(w, statusCode, errResponse{err.Error()})
}
