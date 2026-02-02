package helpers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type HttpRequestHandler struct {
	client *http.Client
	logger *slog.Logger
}

func NewHttpRequestHandler(client *http.Client, logger *slog.Logger) *HttpRequestHandler {
	return &HttpRequestHandler{client: client, logger: logger}
}

func (h *HttpRequestHandler) Request(ctx context.Context, requestData HttpRequestDetails) (string, error) {
	h.logger.DebugContext(ctx, "calling Url", "Url", requestData.Url, "Body", requestData.Body, "Method", requestData.Method)

	requestReader := strings.NewReader(requestData.Body)
	httpRequest, err := http.NewRequest(requestData.Method, requestData.Url, requestReader)
	if err != nil {
		h.logger.ErrorContext(ctx, "error creating Request", "err", err)
		return "", errors.Wrap(err, "error creating Request")
	}

	for k, v := range requestData.Headers {
		httpRequest.Header.Add(k, v)
	}

	resp, err := h.client.Do(httpRequest)
	if err != nil || resp == nil {
		h.logger.ErrorContext(ctx, "error performing Request", "err", err)
		return "", errors.Wrap(err, "error performing Request")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		h.logger.ErrorContext(ctx, "authorisation failed")
		return "", &UnauthorisedError{"authorisation failed"}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.ErrorContext(ctx, "error reading response Body", "err", err)
		return "", errors.Wrap(err, "error reading response Body")
	}

	str := string(data)
	lines := strings.Split(str, "\n")
	if len(lines) < 1 {
		h.logger.ErrorContext(ctx, "no lines found in response Body", "err", err)
		return "", errors.New("no lines found in response Body")
	}

	h.logger.DebugContext(ctx, "got response from Url", "response string", lines[0])
	return lines[0], nil
}

type HttpRequestDetails struct {
	Method  string
	Url     string
	Body    string
	Headers map[string]string
}
