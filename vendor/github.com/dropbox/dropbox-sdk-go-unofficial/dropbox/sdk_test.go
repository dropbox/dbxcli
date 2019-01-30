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

package dropbox_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/users"
)

func generateURL(base string, namespace string, route string) string {
	return fmt.Sprintf("%s/%s/%s", base, namespace, route)
}

func TestInternalError(t *testing.T) {
	eString := "internal server error"
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, eString, http.StatusInternalServerError)
		}))
	defer ts.Close()

	config := dropbox.Config{Client: ts.Client(), LogLevel: dropbox.LogDebug,
		URLGenerator: func(hostType string, style string, namespace string, route string) string {
			return generateURL(ts.URL, namespace, route)
		}}
	client := users.New(config)
	v, e := client.GetCurrentAccount()
	if v != nil || strings.Trim(e.Error(), "\n") != eString {
		t.Errorf("v: %v e: '%s'\n", v, e.Error())
	}
}

func TestRateLimitPlainText(t *testing.T) {
	eString := "too_many_requests"
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Retry-After", "10")
			http.Error(w, eString, http.StatusTooManyRequests)
		}))
	defer ts.Close()

	config := dropbox.Config{Client: ts.Client(), LogLevel: dropbox.LogDebug,
		URLGenerator: func(hostType string, style string, namespace string, route string) string {
			return generateURL(ts.URL, namespace, route)
		}}
	client := users.New(config)
	_, e := client.GetCurrentAccount()
	re, ok := e.(auth.RateLimitAPIError)
	if !ok {
		t.Errorf("Unexpected error type: %T\n", e)
	}
	if re.RateLimitError.RetryAfter != 10 {
		t.Errorf("Unexpected retry-after value: %d\n", re.RateLimitError.RetryAfter)
	}
	if re.RateLimitError.Reason.Tag != auth.RateLimitReasonTooManyRequests {
		t.Errorf("Unexpected reason: %v\n", re.RateLimitError.Reason)
	}
}

func TestRateLimitJSON(t *testing.T) {
	eString := `{"error_summary": "too_many_requests/..", "error": {"reason": {".tag": "too_many_requests"}, "retry_after": 300}}`
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Retry-After", "10")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(eString))
		}))
	defer ts.Close()

	config := dropbox.Config{Client: ts.Client(), LogLevel: dropbox.LogDebug,
		URLGenerator: func(hostType string, style string, namespace string, route string) string {
			return generateURL(ts.URL, namespace, route)
		}}
	client := users.New(config)
	_, e := client.GetCurrentAccount()
	re, ok := e.(auth.RateLimitAPIError)
	if !ok {
		t.Errorf("Unexpected error type: %T\n", e)
	}
	if re.RateLimitError.RetryAfter != 300 {
		t.Errorf("Unexpected retry-after value: %d\n", re.RateLimitError.RetryAfter)
	}
	if re.RateLimitError.Reason.Tag != auth.RateLimitReasonTooManyRequests {
		t.Errorf("Unexpected reason: %v\n", re.RateLimitError.Reason)
	}
}

func TestAuthError(t *testing.T) {
	eString := `{"error_summary": "user_suspended/...", "error": {".tag": "user_suspended"}}`
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(eString))
		}))
	defer ts.Close()

	config := dropbox.Config{Client: ts.Client(), LogLevel: dropbox.LogDebug,
		URLGenerator: func(hostType string, style string, namespace string, route string) string {
			return generateURL(ts.URL, namespace, route)
		}}
	client := users.New(config)
	_, e := client.GetCurrentAccount()
	re, ok := e.(auth.AuthAPIError)
	if !ok {
		t.Errorf("Unexpected error type: %T\n", e)
	}
	fmt.Printf("ERROR is %v\n", re)
	if re.AuthError.Tag != auth.AuthErrorUserSuspended {
		t.Errorf("Unexpected tag: %s\n", re.AuthError.Tag)
	}
}

func TestAccessError(t *testing.T) {
	eString := `{"error_summary": "access_error/...",
	"error": {
		".tag": "paper_access_denied",
	  "paper_access_denied": {".tag": "not_paper_user"}
	}}`
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(eString))
		}))
	defer ts.Close()

	config := dropbox.Config{Client: ts.Client(), LogLevel: dropbox.LogDebug,
		URLGenerator: func(hostType string, style string, namespace string, route string) string {
			return generateURL(ts.URL, namespace, route)
		}}
	client := users.New(config)
	_, e := client.GetCurrentAccount()
	re, ok := e.(auth.AccessAPIError)
	if !ok {
		t.Errorf("Unexpected error type: %T\n", e)
	}
	if re.AccessError.Tag != auth.AccessErrorPaperAccessDenied {
		t.Errorf("Unexpected tag: %s\n", re.AccessError.Tag)
	}
	if re.AccessError.PaperAccessDenied.Tag != auth.PaperAccessErrorNotPaperUser {
		t.Errorf("Unexpected tag: %s\n", re.AccessError.PaperAccessDenied.Tag)
	}
}
