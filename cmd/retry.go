package cmd

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

const (
	maxRetries     = 5
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
)

var retrySleep = time.Sleep

func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	var serverErr auth.ServerError
	if errors.As(err, &serverErr) {
		return true
	}

	var rateLimitErr auth.RateLimitAPIError
	if errors.As(err, &rateLimitErr) {
		return true
	}

	if isTooManyWriteOperationsError(err) {
		return true
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	var sdkErr dropbox.SDKInternalError
	if errors.As(err, &sdkErr) {
		return sdkErr.StatusCode >= 500 && sdkErr.StatusCode <= 599
	}

	var netErr net.Error
	return errors.As(err, &netErr)
}

func isTooManyWriteOperationsError(err error) bool {
	var uploadErr files.UploadAPIError
	if errors.As(err, &uploadErr) && uploadErr.EndpointError != nil {
		endpointErr := uploadErr.EndpointError
		if endpointErr.Tag == files.UploadErrorPath &&
			endpointErr.Path != nil &&
			endpointErr.Path.Reason != nil &&
			endpointErr.Path.Reason.Tag == files.WriteErrorTooManyWriteOperations {
			return true
		}
	}

	var finishErr files.UploadSessionFinishAPIError
	if errors.As(err, &finishErr) && finishErr.EndpointError != nil {
		endpointErr := finishErr.EndpointError
		if endpointErr.Tag == files.UploadSessionFinishErrorTooManyWriteOperations {
			return true
		}
		if endpointErr.Tag == files.UploadSessionFinishErrorPath &&
			endpointErr.Path != nil &&
			endpointErr.Path.Tag == files.WriteErrorTooManyWriteOperations {
			return true
		}
	}

	return false
}

func retryDelay(err error, backoff time.Duration) (time.Duration, bool) {
	var rateLimitErr auth.RateLimitAPIError
	if !errors.As(err, &rateLimitErr) {
		return backoff, false
	}
	if rateLimitErr.RateLimitError == nil {
		return backoff, false
	}
	return time.Duration(rateLimitErr.RateLimitError.RetryAfter) * time.Second, true
}

func retryWithBackoff(fn func() error) error {
	backoff := initialBackoff
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if !isTransientError(lastErr) {
			return lastErr
		}
		if attempt == maxRetries {
			break
		}
		delay, isRateLimit := retryDelay(lastErr, backoff)
		retrySleep(delay)
		if !isRateLimit {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return lastErr
}
