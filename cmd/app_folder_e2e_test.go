package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/users"
)

const appFolderE2EEndpointEnv = "DBXCLI_TEST_APP_FOLDER_ENDPOINT"

// Redirect the subprocess's real SDK clients to a local API simulator.
// Keep token loading, OAuth transport, command execution, and output unchanged.
func installAppFolderHTTPFixture(endpoint string) {
	withEndpoint := func(cfg dropbox.Config) dropbox.Config {
		cfg.URLGenerator = func(_, namespace, route string) string {
			return endpoint + "/2/" + namespace + "/" + route
		}
		return cfg
	}
	filesNewFunc = func(cfg dropbox.Config) filesClient {
		return files.NewContext(withEndpoint(cfg))
	}
	usersNewFunc = func(cfg dropbox.Config) usersClient {
		return users.NewContext(withEndpoint(cfg))
	}
}

func TestAppFolderEndToEnd(t *testing.T) {
	for _, source := range []string{"env", "saved"} {
		for _, scenario := range []struct {
			name, env string
			flags     []string
			wantError bool
		}{
			{name: "CLI flag", env: "false", flags: []string{"--app-folder"}},
			{name: "environment", env: "true"},
			{name: "CLI overrides invalid environment", env: "invalid", flags: []string{"--app-folder"}},
			{name: "default reproduces sandbox error", wantError: true},
			{name: "explicit false restores path root", env: "true", flags: []string{"--app-folder=false"}, wantError: true},
		} {
			t.Run(source+"/"+scenario.name, func(t *testing.T) {
				requests := make(chan string, 8)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests <- r.URL.Path
					if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer app-token" {
						t.Error("expected authenticated SDK POST")
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					switch r.URL.Path {
					case "/2/users/get_current_account":
						_, _ = io.WriteString(w, `{"account_id":"dbid:test","root_info":{".tag":"user","root_namespace_id":"user-root","home_namespace_id":"user-root"}}`)
					case "/2/files/list_folder":
						if r.Header.Get("Dropbox-API-Path-Root") != "" {
							w.WriteHeader(http.StatusBadRequest)
							_, _ = io.WriteString(w, "Path root is not supported for sandbox app")
							return
						}
						var arg files.ListFolderArg
						if err := json.NewDecoder(r.Body).Decode(&arg); err != nil || arg.Path != "" {
							t.Errorf("expected listing at logical app root, got path %q, error %v", arg.Path, err)
							w.WriteHeader(http.StatusBadRequest)
							return
						}
						_, _ = io.WriteString(w, `{"entries":[{".tag":"folder","name":"Reports","id":"id:reports","path_lower":"/reports","path_display":"/Reports"}],"cursor":"cursor","has_more":false}`)
					default:
						t.Errorf("unexpected API endpoint %s", r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer server.Close()

				authFile := filepath.Join(t.TempDir(), "auth.json")
				env := map[string]string{envAccessToken: "", envAuthFile: authFile, envAppFolder: scenario.env, appFolderE2EEndpointEnv: server.URL, listingOutputE2EEnv: ""}
				if source == "env" {
					env[envAccessToken] = "app-token"
				} else if err := os.WriteFile(authFile, []byte(`{"":{"personal":"app-token"}}`), 0600); err != nil {
					t.Fatal(err)
				}
				args := append(append([]string{}, scenario.flags...), "ls", "/", "--output=json")
				// An unset setting exercises the default; empty is an invalid boolean.
				if scenario.env == "" {
					delete(env, envAppFolder)
					unsetEnvForTest(t, envAppFolder)
				}
				code, stdout, stderr := executeExitTestSubprocess(t, args, env)
				if scenario.wantError {
					if code == exitCodeSuccess || !strings.Contains(stdout, "Path root is not supported for sandbox app") {
						t.Fatalf("expected sandbox rejection, exit = %d, stdout = %s, stderr = %s", code, stdout, stderr)
					}
				} else if code != exitCodeSuccess || stderr != "" || !strings.Contains(stdout, `"path_display":"/Reports"`) {
					t.Fatalf("listing failed: exit = %d, stdout = %s, stderr = %s", code, stdout, stderr)
				}
				wantCalls := []string{"/2/files/list_folder"}
				if scenario.wantError {
					wantCalls = append([]string{"/2/users/get_current_account"}, wantCalls...)
				}
				close(requests)
				var gotCalls []string
				for route := range requests {
					gotCalls = append(gotCalls, route)
				}
				if strings.Join(gotCalls, ",") != strings.Join(wantCalls, ",") {
					t.Fatalf("API calls = %v, want %v", gotCalls, wantCalls)
				}
			})
		}
	}
}
