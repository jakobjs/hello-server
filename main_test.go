package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthCheckerIsHealthy(t *testing.T) {
	startedAt := time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name    string
		now     time.Time
		healthy bool
	}{
		{
			name:    "healthy before first failure window",
			now:     startedAt.Add(29*time.Minute + 59*time.Second),
			healthy: true,
		},
		{
			name:    "unhealthy at first failure start",
			now:     startedAt.Add(30 * time.Minute),
			healthy: false,
		},
		{
			name:    "unhealthy within first failure window",
			now:     startedAt.Add(37 * time.Minute),
			healthy: false,
		},
		{
			name:    "healthy after first failure window",
			now:     startedAt.Add(40 * time.Minute),
			healthy: true,
		},
		{
			name:    "unhealthy in later failure window",
			now:     startedAt.Add(92 * time.Minute),
			healthy: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			checker := newHealthChecker(startedAt)
			checker.now = func() time.Time { return testCase.now }

			if got := checker.isHealthy(); got != testCase.healthy {
				t.Fatalf("isHealthy() = %v, want %v", got, testCase.healthy)
			}
		})
	}
}

func TestHealthServer(t *testing.T) {
	startedAt := time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)
	checker := newHealthChecker(startedAt)
	checker.now = func() time.Time { return startedAt.Add(35 * time.Minute) }

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	HealthServer(checker).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func TestReadinessCheckerIsReady(t *testing.T) {
	startedAt := time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name  string
		now   time.Time
		ready bool
	}{
		{
			name:  "not ready before threshold",
			now:   startedAt.Add(119 * time.Second),
			ready: false,
		},
		{
			name:  "ready at threshold",
			now:   startedAt.Add(2 * time.Minute),
			ready: true,
		},
		{
			name:  "ready after threshold",
			now:   startedAt.Add(3 * time.Minute),
			ready: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			checker := newReadinessChecker(startedAt)
			checker.now = func() time.Time { return testCase.now }

			if got := checker.isReady(); got != testCase.ready {
				t.Fatalf("isReady() = %v, want %v", got, testCase.ready)
			}
		})
	}
}

func TestReadyServer(t *testing.T) {
	startedAt := time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)
	testCases := []struct {
		name       string
		now        time.Time
		wantStatus int
		wantBody   string
	}{
		{
			name:       "returns not ready before threshold",
			now:        startedAt.Add(90 * time.Second),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "not ready\n",
		},
		{
			name:       "returns ready after threshold",
			now:        startedAt.Add(2*time.Minute + 1*time.Second),
			wantStatus: http.StatusOK,
			wantBody:   "ready\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			checker := newReadinessChecker(startedAt)
			checker.now = func() time.Time { return testCase.now }

			request := httptest.NewRequest(http.MethodGet, "/ready", nil)
			recorder := httptest.NewRecorder()

			ReadyServer(checker).ServeHTTP(recorder, request)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status code = %d, want %d", recorder.Code, testCase.wantStatus)
			}

			if recorder.Body.String() != testCase.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), testCase.wantBody)
			}
		})
	}
}
