package orchestration

import (
	"context"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/helpers"
)

type MockHttpHelper struct {
	RequestData    []helpers.HttpRequestDetails
	RequestResults []string
	RequestErrors  []error
	RequestCalled  []bool
	RequestCount   int
}

func newMockHttpHelper() *MockHttpHelper {
	return &MockHttpHelper{
		RequestData:    []helpers.HttpRequestDetails{},
		RequestResults: []string{},
		RequestErrors:  []error{},
		RequestCount:   0,
	}
}

func (m *MockHttpHelper) Request(_ context.Context, requestData helpers.HttpRequestDetails) (string, error) {
	m.RequestData = append(m.RequestData, requestData)
	defer func() {
		m.RequestCount++
	}()
	return m.RequestResults[m.RequestCount], m.RequestErrors[m.RequestCount]
}
