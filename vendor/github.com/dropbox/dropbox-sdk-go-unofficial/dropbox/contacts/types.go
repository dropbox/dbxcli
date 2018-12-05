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

// Package contacts : has no documentation (yet)
package contacts

import (
	"encoding/json"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
)

// DeleteManualContactsArg : has no documentation (yet)
type DeleteManualContactsArg struct {
	// EmailAddresses : List of manually added contacts to be deleted.
	EmailAddresses []string `json:"email_addresses"`
}

// NewDeleteManualContactsArg returns a new DeleteManualContactsArg instance
func NewDeleteManualContactsArg(EmailAddresses []string) *DeleteManualContactsArg {
	s := new(DeleteManualContactsArg)
	s.EmailAddresses = EmailAddresses
	return s
}

// DeleteManualContactsError : has no documentation (yet)
type DeleteManualContactsError struct {
	dropbox.Tagged
	// ContactsNotFound : Can't delete contacts from this list. Make sure the
	// list only has manually added contacts. The deletion was cancelled.
	ContactsNotFound []string `json:"contacts_not_found,omitempty"`
}

// Valid tag values for DeleteManualContactsError
const (
	DeleteManualContactsErrorContactsNotFound = "contacts_not_found"
	DeleteManualContactsErrorOther            = "other"
)

// UnmarshalJSON deserializes into a DeleteManualContactsError instance
func (u *DeleteManualContactsError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
		// ContactsNotFound : Can't delete contacts from this list. Make sure
		// the list only has manually added contacts. The deletion was
		// cancelled.
		ContactsNotFound json.RawMessage `json:"contacts_not_found,omitempty"`
	}
	var w wrap
	var err error
	if err = json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "contacts_not_found":
		err = json.Unmarshal(body, &u.ContactsNotFound)

		if err != nil {
			return err
		}
	}
	return nil
}
