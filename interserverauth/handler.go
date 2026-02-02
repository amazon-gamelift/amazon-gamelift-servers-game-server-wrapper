package interserverauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/helpers"
	"github.com/pkg/errors"
)

const (
	authHeaderKey = "Authorization"
)

type Handler struct {
	clientId         string
	clientSecret     string
	tokenEndpoint    string
	logger           *slog.Logger
	requestHelper    requestHelper
	token            *string
	tokenMutex       *sync.Mutex
	authHeaderPrefix string
}

func New(authHeaderPrefix string, clientId string, clientSecret string, tokenEndpoint string, requestHelper requestHelper, logger *slog.Logger) *Handler {
	return &Handler{
		clientId:         clientId,
		clientSecret:     clientSecret,
		tokenEndpoint:    tokenEndpoint,
		logger:           logger,
		authHeaderPrefix: authHeaderPrefix,
		requestHelper:    requestHelper,
		token:            nil,
		tokenMutex:       &sync.Mutex{},
	}
}

func (a *Handler) AcquireAccessToken(ctx context.Context) error {
	a.tokenMutex.Lock()
	defer a.tokenMutex.Unlock()

	encodedId := base64.StdEncoding.EncodeToString([]byte(a.clientId))
	encodedSecret := base64.StdEncoding.EncodeToString([]byte(a.clientSecret))
	request := helpers.HttpRequestDetails{
		Method:  http.MethodGet,
		Url:     a.tokenEndpoint,
		Headers: map[string]string{authHeaderKey: fmt.Sprintf("%s %s:%s", a.authHeaderPrefix, encodedId, encodedSecret)},
	}

	token, err := a.requestHelper.Request(ctx, request)
	if err != nil {
		msg := fmt.Sprintf("Error acquiring token: %v", err)
		a.logger.ErrorContext(ctx, msg)
		return errors.Wrap(err, msg)
	}

	a.logger.DebugContext(ctx, "acquired token")
	a.token = &token

	if !a.isValidToken(ctx) {
		a.logger.ErrorContext(ctx, "deleting invalid token")
		a.token = nil
		return errors.New("an invalid token was acquired")
	}

	return nil
}

func (a *Handler) IsValidAccessToken(ctx context.Context) bool {
	a.tokenMutex.Lock()
	defer a.tokenMutex.Unlock()

	return a.isValidToken(ctx)
}

func (a *Handler) isValidToken(ctx context.Context) bool {
	a.logger.DebugContext(ctx, "checking auth token is valid")

	if a.token == nil || len(*a.token) == 0 {
		a.logger.DebugContext(ctx, "token not valid: token is empty")
		return false
	}

	tokenParts := strings.Split(*a.token, ".")
	if len(tokenParts) != 3 {
		a.logger.ErrorContext(ctx, "token not valid: token does not contain 3 parts")
		return false
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(tokenParts[1])
	if err != nil {
		a.logger.ErrorContext(ctx, "token not valid: base64 payload decoding failed")
		return false
	}

	payload := JwtPayload{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		a.logger.ErrorContext(ctx, "token not valid: json payload decoding failed")
		return false
	}

	tokenExpiry := time.Unix(payload.ExpirationTime, 0)
	if time.Now().In(time.UTC).After(tokenExpiry) {
		a.logger.ErrorContext(ctx, "token not valid: token expired")
		return false
	}

	return true
}

func (a *Handler) GetAuthHeader(ctx context.Context) string {
	a.tokenMutex.Lock()
	defer a.tokenMutex.Unlock()

	a.logger.DebugContext(ctx, "getting auth token")
	return fmt.Sprintf("Bearer %s", *a.token)
}

type requestHelper interface {
	Request(ctx context.Context, requestData helpers.HttpRequestDetails) (string, error)
}
