package main

import (
	"fmt"
	"net/url"
)

const (
	dropboxScheme = "dropbox"
)

func parseDropboxUri(uri string) (path string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}

	if u.Scheme != dropboxScheme {
		err = fmt.Errorf("Path should start with %s://", dropboxScheme)
		return
	}

	if len(u.Host) == 0 && len(u.Path) == 0 {
		return
	}

	path = fmt.Sprintf("%s%s", u.Host, u.Path)

	if path[0] != '/' {
		path = fmt.Sprintf("/%s", path)
	}

	if path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}

	return
}
