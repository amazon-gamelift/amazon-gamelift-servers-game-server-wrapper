package orchestration

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/helpers"
	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/v5/model"
	"github.com/stretchr/testify/assert"
)

func TestSender_TokenEmpty(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.RequestHelper.RequestResults = []string{"", ""}

	// Act
	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)

	// Assert
	assert.NotNil(t, err, fmt.Sprintf("Expected error on start game session"))
	assert.Contains(t, err.Error(), "Error acquiring token: token is empty")
}

func TestSender_ErrorOnGettingToken(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.RequestHelper.RequestResults = []string{"", ""}
	h.RequestHelper.RequestErrors = []error{errors.New("test-error"), nil}

	// Act
	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)

	// Assert
	assert.NotNil(t, err, fmt.Sprintf("Expected error on start game session"))
	assert.Contains(t, err.Error(), "Error acquiring token: test-error")
}

func TestSender_OnStartGameSession_HappyPath(t *testing.T) {
	// Arrange
	h := newTestHelper()

	// Act
	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	assert.Contains(t, h.LogBuffer.String(), "emitting start game session event to the orchestration layer")
	assert.Equal(t, 2, h.RequestHelper.RequestCount, "request was not called the correct number of times")
	assert.Equal(t, http.MethodGet, h.RequestHelper.RequestData[0].Method, "the wrong request method was used")
	assert.Equal(t, h.Config.Method, h.RequestHelper.RequestData[1].Method, "the wrong request method was used")
	assert.Equal(t, h.Config.GetTokenUrl, h.RequestHelper.RequestData[0].Url, "the wrong request url was used")
	assert.Equal(t, h.Config.Url, h.RequestHelper.RequestData[1].Url, "the wrong request url was used")
	assert.Equal(t, h.Config.HeaderValue, h.RequestHelper.RequestData[1].Headers[h.Config.HeaderKey], "the wrong request header was used")

	assert.NotEmpty(t, h.RequestHelper.RequestData[1].Body, "request body is empty")
	assert.Contains(t, h.RequestHelper.RequestData[1].Body, "\"Type\":\"PlacementActive\"")
	assert.Contains(t, h.RequestHelper.RequestData[1].Body, "\"Id\":\"test-game-session-id\"")
	assert.Contains(t, h.RequestHelper.RequestData[1].Body, "\"PlacementId\":\"test-game-session-id\"")

	encoded := base64.StdEncoding.EncodeToString([]byte("test-client-id:test-client-secret"))
	assert.Contains(t, h.RequestHelper.RequestData[0].Headers["Authorization"], fmt.Sprintf("Basic %s", encoded))
	assert.NotEmpty(t, h.RequestHelper.RequestData[1].Headers["Authorization"], "auth header is empty")
	assert.Equal(t, "Basic test-token", h.RequestHelper.RequestData[1].Headers["Authorization"], "auth header has wrong value")
}

func TestSender_OnStartGameSession_HappyPath_DontEmit(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.Config.EmitCustomEvents = false

	// Act
	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	assert.Contains(t, h.LogBuffer.String(), "emitting events to the orchestration layer is disabled")
	assert.Equal(t, 0, h.RequestHelper.RequestCount, "request was called when it should not have been")
}

func TestSender_OnStartGameSession_ExpiredToken_HappyPath(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.RequestHelper.RequestResults = []string{"test-token-one", "", "test-token-two", ""}
	h.RequestHelper.RequestErrors = []error{nil, &helpers.UnauthorisedError{Err: "test-unauthorised-error"}, nil, nil}

	// Act
	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	assert.Contains(t, h.LogBuffer.String(), "emitting start game session event to the orchestration layer")
	assert.Equal(t, 4, h.RequestHelper.RequestCount, "request was not called the correct number of times")
	assert.Equal(t, http.MethodGet, h.RequestHelper.RequestData[0].Method, "the wrong request method was used")
	assert.Equal(t, http.MethodGet, h.RequestHelper.RequestData[2].Method, "the wrong request method was used")
	assert.Equal(t, h.Config.GetTokenUrl, h.RequestHelper.RequestData[0].Url, "the wrong request url was used")
	assert.Equal(t, h.Config.GetTokenUrl, h.RequestHelper.RequestData[2].Url, "the wrong request url was used")

	encoded := base64.StdEncoding.EncodeToString([]byte("test-client-id:test-client-secret"))
	assert.Contains(t, h.RequestHelper.RequestData[0].Headers["Authorization"], fmt.Sprintf("Basic %s", encoded))
	assert.Contains(t, h.RequestHelper.RequestData[2].Headers["Authorization"], fmt.Sprintf("Basic %s", encoded))
	assert.NotEmpty(t, h.RequestHelper.RequestData[3].Headers["Authorization"], "auth header is empty")
	assert.Equal(t, "Basic test-token-two", h.RequestHelper.RequestData[3].Headers["Authorization"], "auth header has wrong value")
}

func TestSender_OnStartGameSession_ExpiredToken_ErrorOnRefreshingToken(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.RequestHelper.RequestResults = []string{"test-token-one", "", "", ""}
	h.RequestHelper.RequestErrors = []error{nil, &helpers.UnauthorisedError{Err: "test-unauthorised-error"}, errors.New("test-error"), nil}

	// Act
	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)

	// Assert
	assert.NotNil(t, err, fmt.Sprintf("Expected error on start game session"))
	assert.Contains(t, err.Error(), "Error re-acquiring token")
	assert.Contains(t, err.Error(), "test-error")
}

func TestSender_OnStartGameSession_RequestFailure(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.RequestHelper.RequestErrors = []error{nil, errors.New("test-error")}

	// Act
	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)

	// Assert
	assert.NotNil(t, err, "Error was not returned")
	assert.Contains(t, err.Error(), "test-error")
	assert.Contains(t, err.Error(), "Error emitting event")
	assert.Contains(t, err.Error(), "Error emitting OnStartGameSession event")
	assert.Contains(t, h.LogBuffer.String(), "Error emitting event")
	assert.Contains(t, h.LogBuffer.String(), "Error emitting OnStartGameSession event")
}

func TestSender_OnHostingTerminate_HappyPath(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.RequestHelper.RequestResults = []string{"test-token", "", ""}
	h.RequestHelper.RequestErrors = []error{nil, nil, nil}

	// Act
	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	h.LogBuffer.Reset()

	err = h.Sender.OnHostingTerminate(*h.Ctx)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on hosting terminate: %v", err))
	assert.Contains(t, h.LogBuffer.String(), "emitting hosting terminate event to the orchestration layer")
	assert.Equal(t, 3, h.RequestHelper.RequestCount, "request was not called the correct number of times")
	assert.Equal(t, h.Config.Method, h.RequestHelper.RequestData[2].Method, "the wrong request method was used")
	assert.Equal(t, h.Config.Url, h.RequestHelper.RequestData[2].Url, "the wrong request url was used")
	assert.Equal(t, h.Config.HeaderValue, h.RequestHelper.RequestData[2].Headers[h.Config.HeaderKey], "the wrong request header was used")
	assert.Equal(t, h.Config.HeaderValue, h.RequestHelper.RequestData[2].Headers[h.Config.HeaderKey], "the wrong request header was used")

	assert.NotEmpty(t, h.RequestHelper.RequestData[2].Body, "request body is empty")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"Type\":\"PlacementTerminated\"")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"Account\":\"test-account\"")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"PlacementId\":\"test-game-session-id\"")

	assert.NotEmpty(t, h.RequestHelper.RequestData[2].Headers["Authorization"], "auth header is empty")
	assert.Equal(t, "Basic test-token", h.RequestHelper.RequestData[2].Headers["Authorization"], "auth header has wrong value")
}

func TestSender_OnHostingTerminate_HappyPath_DontEmit(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.Config.EmitCustomEvents = false

	// Act
	err := h.Sender.OnHostingTerminate(*h.Ctx)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on hosting terminate: %v", err))
	assert.Contains(t, h.LogBuffer.String(), "emitting events to the orchestration layer is disabled")
	assert.Equal(t, 0, h.RequestHelper.RequestCount, "request was called when it should not have been")
}

func TestSender_OnHostingTerminate_RequestFailure(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.RequestHelper.RequestResults = []string{"test-token", "", ""}
	h.RequestHelper.RequestErrors = []error{nil, nil, errors.New("test-error")}

	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	h.LogBuffer.Reset()

	// Act
	err = h.Sender.OnHostingTerminate(*h.Ctx)

	// Assert
	assert.NotNil(t, err, "Error was not returned")
	assert.Contains(t, err.Error(), "test-error")
	assert.Contains(t, err.Error(), "Error emitting event")
	assert.Contains(t, err.Error(), "Error emitting OnHostingTerminate event")
	assert.Contains(t, h.LogBuffer.String(), "Error emitting event")
	assert.Contains(t, h.LogBuffer.String(), "Error emitting OnHostingTerminate event")
}

func TestSender_OnHealthCheck_HappyPath(t *testing.T) {
	// Arrange
	h := newTestHelper()

	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	h.LogBuffer.Reset()

	// Act
	err = h.Sender.OnHealthCheck(*h.Ctx)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on health check: %v", err))

	// TODO: Check if any behaviour here needs defined.
}

func TestSender_OnHealthCheck_HappyPath_DontEmit(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.Config.EmitCustomEvents = false

	// Act
	err := h.Sender.OnHealthCheck(*h.Ctx)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on hosting terminate: %v", err))
	assert.Contains(t, h.LogBuffer.String(), "emitting events to the orchestration layer is disabled")
	assert.Equal(t, 0, h.RequestHelper.RequestCount, "request was called when it should not have been")
}

func TestSender_OnUpdateGameSession_HappyPath_PlacementActive(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.RequestHelper.RequestResults = []string{"test-token", "", ""}
	h.RequestHelper.RequestErrors = []error{nil, nil, nil}

	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	h.LogBuffer.Reset()

	// Act
	err = h.Sender.OnUpdateGameSession(*h.Ctx, *h.UpdateGameSession)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on update game session: %v", err))
	assert.Contains(t, h.LogBuffer.String(), "emitting update game session event to the orchestration layer")
	assert.Equal(t, 3, h.RequestHelper.RequestCount, "request was not called the correct number of times")
	assert.Equal(t, h.Config.Method, h.RequestHelper.RequestData[2].Method, "the wrong request method was used")
	assert.Equal(t, h.Config.Url, h.RequestHelper.RequestData[2].Url, "the wrong request url was used")
	assert.Equal(t, h.Config.HeaderValue, h.RequestHelper.RequestData[2].Headers[h.Config.HeaderKey], "the wrong request header was used")

	assert.NotEmpty(t, h.RequestHelper.RequestData[2].Body, "request body is empty")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"Type\":\"PlacementActive\"")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"Id\":\"test-game-session-id\"")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"PlacementId\":\"test-game-session-id\"")

	assert.NotEmpty(t, h.RequestHelper.RequestData[2].Headers["Authorization"], "auth header is empty")
	assert.Equal(t, "Basic test-token", h.RequestHelper.RequestData[2].Headers["Authorization"], "auth header has wrong value")
}

func TestSender_OnUpdateGameSession_HappyPath_PlacementActive_NoDormantProperty(t *testing.T) {
	// Arrange
	h := newTestHelper()
	delete(h.GameSession.GameProperties, "dormant")
	h.RequestHelper.RequestResults = []string{"test-token", "", ""}
	h.RequestHelper.RequestErrors = []error{nil, nil, nil}

	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	h.LogBuffer.Reset()

	// Act
	err = h.Sender.OnUpdateGameSession(*h.Ctx, *h.UpdateGameSession)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on update game session: %v", err))
	assert.Contains(t, h.LogBuffer.String(), "emitting update game session event to the orchestration layer")
	assert.Contains(t, h.LogBuffer.String(), "game property 'dormant' is missing")
	assert.Equal(t, 3, h.RequestHelper.RequestCount, "request was not called the correct number of times")
	assert.Equal(t, h.Config.Method, h.RequestHelper.RequestData[2].Method, "the wrong request method was used")
	assert.Equal(t, h.Config.Url, h.RequestHelper.RequestData[2].Url, "the wrong request url was used")
	assert.Equal(t, h.Config.HeaderValue, h.RequestHelper.RequestData[2].Headers[h.Config.HeaderKey], "the wrong request header was used")

	assert.NotEmpty(t, h.RequestHelper.RequestData[2].Body, "request body is empty")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"Type\":\"PlacementActive\"")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"Id\":\"test-game-session-id\"")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"PlacementId\":\"test-game-session-id\"")

	assert.NotEmpty(t, h.RequestHelper.RequestData[2].Headers["Authorization"], "auth header is empty")
	assert.Equal(t, "Basic test-token", h.RequestHelper.RequestData[2].Headers["Authorization"], "auth header has wrong value")
}

func TestSender_OnUpdateGameSession_HappyPath_PlacementDormant(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.GameSession.GameProperties["dormant"] = "true"
	h.RequestHelper.RequestResults = []string{"test-token", "", ""}
	h.RequestHelper.RequestErrors = []error{nil, nil, nil}

	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	h.LogBuffer.Reset()

	// Act
	err = h.Sender.OnUpdateGameSession(*h.Ctx, *h.UpdateGameSession)

	// Assert
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on update game session: %v", err))
	assert.Contains(t, h.LogBuffer.String(), "emitting update game session event to the orchestration layer")
	assert.Equal(t, 3, h.RequestHelper.RequestCount, "request was not called the correct number of times")
	assert.Equal(t, h.Config.Method, h.RequestHelper.RequestData[2].Method, "the wrong request method was used")
	assert.Equal(t, h.Config.Url, h.RequestHelper.RequestData[2].Url, "the wrong request url was used")
	assert.Equal(t, h.Config.HeaderValue, h.RequestHelper.RequestData[2].Headers[h.Config.HeaderKey], "the wrong request header was used")

	assert.NotEmpty(t, h.RequestHelper.RequestData[2].Body, "request body is empty")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"Type\":\"PlacementDormant\"")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"Id\":\"test-game-session-id\"")
	assert.Contains(t, h.RequestHelper.RequestData[2].Body, "\"PlacementId\":\"test-game-session-id\"")

	assert.NotEmpty(t, h.RequestHelper.RequestData[2].Headers["Authorization"], "auth header is empty")
	assert.Equal(t, "Basic test-token", h.RequestHelper.RequestData[2].Headers["Authorization"], "auth header has wrong value")
}

func TestSender_OnUpdateGameSession_RequestFailure(t *testing.T) {
	// Arrange
	h := newTestHelper()
	h.RequestHelper.RequestResults = []string{"test-token", "", ""}
	h.RequestHelper.RequestErrors = []error{nil, nil, errors.New("test-error")}

	err := h.Sender.OnStartGameSession(*h.Ctx, *h.GameSession)
	assert.Nil(t, err, fmt.Sprintf("Unexpected error on start game session: %v", err))
	h.LogBuffer.Reset()

	// Act
	err = h.Sender.OnUpdateGameSession(*h.Ctx, *h.UpdateGameSession)

	// Assert
	assert.NotNil(t, err, "Expected error on update game session")
	assert.Contains(t, err.Error(), "test-error")
	assert.Contains(t, err.Error(), "Error emitting event")
	assert.Contains(t, err.Error(), "Error emitting OnUpdateGameSession event")
	assert.Contains(t, h.LogBuffer.String(), "Error emitting event")
	assert.Contains(t, h.LogBuffer.String(), "Error emitting OnUpdateGameSession event")
}

func newTestHelper() *testHelper {
	ctx := context.Background()
	logBuffer := bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	reqHelper := newMockHttpHelper()
	reqHelper.RequestErrors = []error{nil, nil}
	reqHelper.RequestResults = []string{"test-token", ""}

	region := "test-region"
	cfg := &config.Orchestration{
		EmitCustomEvents: true,
		Account:          "test-account",
		Resources:        []string{"resource-1", "resource-2", "resource-3"},
		Method:           "test-method",
		Url:              "test-url",
		HeaderKey:        "test-header-key",
		HeaderValue:      "test-header-value",
		GetTokenUrl:      "test-token-url",
		ClientId:         "test-client-id",
		ClientSecret:     "test-client-secret",
	}
	sender := NewSender(logger, reqHelper, cfg, region)
	gameSession := &model.GameSession{
		GameSessionID:             "test-game-session-id",
		GameSessionData:           "test-game-session-data",
		Name:                      "test-game-session-name",
		MatchmakerData:            "test-game-matchmaker-data",
		FleetID:                   "test-fleet-id",
		Location:                  "test-location",
		MaximumPlayerSessionCount: 42,
		IPAddress:                 "test-ip-address",
		Port:                      4242,
		DNSName:                   "test-dns-name",
		GameProperties: map[string]string{
			"zoneId":  "test-zone-id",
			"dormant": "false",
		},
		Status:       nil,
		StatusReason: "test-status-reason",
	}

	reason := model.MatchmakingDataUpdated
	updateGameSession := &model.UpdateGameSession{
		BackfillTicketID: "test-backfill-ticket-id",
		GameSession:      *gameSession,
		UpdateReason:     &reason,
	}

	return &testHelper{
		Sender:            sender,
		Ctx:               &ctx,
		Logger:            logger,
		LogBuffer:         &logBuffer,
		RequestHelper:     reqHelper,
		Config:            cfg,
		Region:            region,
		GameSession:       gameSession,
		UpdateGameSession: updateGameSession,
	}
}

type testHelper struct {
	Sender            *Sender
	Ctx               *context.Context
	Logger            *slog.Logger
	LogBuffer         *bytes.Buffer
	RequestHelper     *MockHttpHelper
	Config            *config.Orchestration
	Region            string
	GameSession       *model.GameSession
	UpdateGameSession *model.UpdateGameSession
}
