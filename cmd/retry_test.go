package cmd

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func stubRetrySleep(t *testing.T) *[]time.Duration {
	t.Helper()

	var delays []time.Duration
	origRetrySleep := retrySleep
	retrySleep = func(delay time.Duration) {
		delays = append(delays, delay)
	}
	t.Cleanup(func() {
		retrySleep = origRetrySleep
	})

	return &delays
}

func writeError(tag string) *files.WriteError {
	return &files.WriteError{Tagged: dropbox.Tagged{Tag: tag}}
}

func uploadWriteThrottleError() files.UploadAPIError {
	return files.UploadAPIError{
		EndpointError: &files.UploadError{
			Tagged: dropbox.Tagged{Tag: files.UploadErrorPath},
			Path: files.NewUploadWriteFailed(
				writeError(files.WriteErrorTooManyWriteOperations),
				"session123",
			),
		},
	}
}

func uploadWriteConflictError() files.UploadAPIError {
	return files.UploadAPIError{
		EndpointError: &files.UploadError{
			Tagged: dropbox.Tagged{Tag: files.UploadErrorPath},
			Path: files.NewUploadWriteFailed(
				writeError(files.WriteErrorConflict),
				"session123",
			),
		},
	}
}

func finishTooManyWriteOperationsError() files.UploadSessionFinishAPIError {
	return files.UploadSessionFinishAPIError{
		EndpointError: &files.UploadSessionFinishError{
			Tagged: dropbox.Tagged{Tag: files.UploadSessionFinishErrorTooManyWriteOperations},
		},
	}
}

func TestIsTransientError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"permanent error", errors.New("path/not_found"), false},
		{"auth error", errors.New("invalid_access_token"), false},
		{
			"server error 5xx",
			auth.ServerError{APIError: dropbox.APIError{ErrorSummary: "internal"}},
			true,
		},
		{
			"rate limit error",
			auth.RateLimitAPIError{RateLimitError: auth.NewRateLimitError(nil)},
			true,
		},
		{
			"sdk internal 500",
			dropbox.SDKInternalError{StatusCode: 500, Content: "server error"},
			true,
		},
		{
			"sdk internal 503",
			dropbox.SDKInternalError{StatusCode: 503, Content: "unavailable"},
			true,
		},
		{
			"sdk internal 400 (not transient)",
			dropbox.SDKInternalError{StatusCode: 400, Content: "bad request"},
			false,
		},
		{
			"net error",
			&net.OpError{Op: "read", Net: "tcp", Err: errors.New("reset")},
			true,
		},
		{
			"single-shot upload too many write operations",
			uploadWriteThrottleError(),
			true,
		},
		{
			"upload session finish too many write operations",
			finishTooManyWriteOperationsError(),
			true,
		},
		{
			"single-shot upload write conflict",
			uploadWriteConflictError(),
			false,
		},
		{
			"unexpected EOF",
			io.ErrUnexpectedEOF,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTransientError(tt.err)
			if got != tt.want {
				t.Errorf("isTransientError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestRetryWithBackoff_Success(t *testing.T) {
	calls := 0
	err := retryWithBackoff(func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryWithBackoff_TransientThenSuccess(t *testing.T) {
	delays := stubRetrySleep(t)
	calls := 0
	err := retryWithBackoff(func() error {
		calls++
		if calls < 3 {
			return auth.ServerError{APIError: dropbox.APIError{ErrorSummary: "500"}}
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
	wantDelays := []time.Duration{initialBackoff, 2 * initialBackoff}
	if len(*delays) != len(wantDelays) {
		t.Fatalf("expected %d sleeps, got %d", len(wantDelays), len(*delays))
	}
	for i, want := range wantDelays {
		if (*delays)[i] != want {
			t.Errorf("sleep %d = %v, want %v", i, (*delays)[i], want)
		}
	}
}

func TestRetryWithBackoff_RateLimitUsesRetryAfter(t *testing.T) {
	delays := stubRetrySleep(t)
	rateLimit := auth.NewRateLimitError(nil)
	rateLimit.RetryAfter = 7

	calls := 0
	err := retryWithBackoff(func() error {
		calls++
		if calls == 1 {
			return auth.RateLimitAPIError{RateLimitError: rateLimit}
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
	if len(*delays) != 1 {
		t.Fatalf("expected 1 sleep, got %d", len(*delays))
	}
	if (*delays)[0] != 7*time.Second {
		t.Errorf("sleep = %v, want %v", (*delays)[0], 7*time.Second)
	}
}

func TestRetryWithBackoff_PermanentError(t *testing.T) {
	calls := 0
	err := retryWithBackoff(func() error {
		calls++
		return errors.New("invalid_access_token")
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retry for permanent error), got %d", calls)
	}
}

func TestRetryWithBackoff_ExhaustsRetries(t *testing.T) {
	delays := stubRetrySleep(t)
	calls := 0
	err := retryWithBackoff(func() error {
		calls++
		return auth.ServerError{APIError: dropbox.APIError{ErrorSummary: "500"}}
	})

	if err == nil {
		t.Error("expected error after exhausting retries, got nil")
	}
	if calls != maxRetries+1 {
		t.Errorf("expected %d calls, got %d", maxRetries+1, calls)
	}
	wantDelays := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
	}
	if len(*delays) != len(wantDelays) {
		t.Fatalf("expected %d sleeps, got %d", len(wantDelays), len(*delays))
	}
	for i, want := range wantDelays {
		if (*delays)[i] != want {
			t.Errorf("sleep %d = %v, want %v", i, (*delays)[i], want)
		}
	}
}
