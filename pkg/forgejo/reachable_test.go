// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.b4mad.industries/agentic-forges/forgejo-mcp/v3/pkg/flag"
)

// A server without a token of its own can't call any API on a Forgejo that
// requires sign-in (REQUIRE_SIGNIN_VIEW), /version included. Its answer still
// proves the instance is there.
func TestVerifyReachable(t *testing.T) {
	cases := []struct {
		name   string
		status int
		ok     bool
	}{
		{"open instance", http.StatusOK, true},
		{"sign-in required", http.StatusForbidden, true},
		{"unauthorized", http.StatusUnauthorized, true},
		{"not Forgejo", http.StatusNotFound, false},
		{"broken", http.StatusInternalServerError, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var auth string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/version" {
					t.Errorf("path = %s", r.URL.Path)
				}
				auth = r.Header.Get("Authorization")
				w.WriteHeader(c.status)
				_, _ = w.Write([]byte(`{"message":"Only signed in user is allowed to call APIs."}`))
			}))
			defer srv.Close()
			flag.URL = srv.URL
			err := VerifyReachable(context.Background())
			if (err == nil) != c.ok {
				t.Fatalf("err = %v, want ok = %v", err, c.ok)
			}
			if auth != "" {
				t.Errorf("sent Authorization %q, want none", auth)
			}
		})
	}
}

func TestVerifyReachableWithoutAServer(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	flag.URL = srv.URL
	srv.Close()
	if err := VerifyReachable(context.Background()); err == nil {
		t.Fatal("want an error when nothing listens")
	}
}
