// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package auth : has no documentation (yet)
package auth

import "github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"

// AuthError : Errors occurred during authentication.
type AuthError struct {
	dropbox.Tagged
}

// Valid tag values for AuthError
const (
	AuthErrorInvalidAccessToken = "invalid_access_token"
	AuthErrorInvalidSelectUser  = "invalid_select_user"
	AuthErrorInvalidSelectAdmin = "invalid_select_admin"
	AuthErrorOther              = "other"
)

// RateLimitError : Error occurred because the app is being rate limited.
type RateLimitError struct {
	// Reason : The reason why the app is being rate limited.
	Reason *RateLimitReason `json:"reason"`
	// RetryAfter : The number of seconds that the app should wait before making
	// another request.
	RetryAfter uint64 `json:"retry_after"`
}

// NewRateLimitError returns a new RateLimitError instance
func NewRateLimitError(Reason *RateLimitReason) *RateLimitError {
	s := new(RateLimitError)
	s.Reason = Reason
	s.RetryAfter = 1
	return s
}

// RateLimitReason : has no documentation (yet)
type RateLimitReason struct {
	dropbox.Tagged
}

// Valid tag values for RateLimitReason
const (
	RateLimitReasonTooManyRequests        = "too_many_requests"
	RateLimitReasonTooManyWriteOperations = "too_many_write_operations"
	RateLimitReasonOther                  = "other"
)
