// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"git.b4mad.industries/agentic-forges/forgejo-mcp/v3/pkg/flag"
	"git.b4mad.industries/agentic-forges/forgejo-mcp/v3/pkg/log"
)

// VerifyReachable checks at startup that the Forgejo instance at flag.URL
// answers, without a credential. It is for a server that has no token of its
// own (every request brings one), which VerifyConnection can't serve: the SDK
// client it builds asks for /version, and a Forgejo that requires sign-in
// (REQUIRE_SIGNIN_VIEW) refuses every API call without a token, that one
// included. So 200, 401 and 403 all prove the instance is there; a network
// error or any other status still stops the server, as a wrong --url should.
func VerifyReachable(ctx context.Context) error {
	url := strings.TrimRight(flag.URL, "/") + "/api/v1/version"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("check Forgejo at %s: %w", flag.URL, err)
	}
	req.Header.Set("User-Agent", userAgent())
	resp, err := rawHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("reach Forgejo at %s: %w", flag.URL, err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	switch resp.StatusCode {
	case http.StatusOK, http.StatusUnauthorized, http.StatusForbidden:
		log.Info("Forgejo instance reachable; each request brings its own token",
			log.SanitizedURLField("url", flag.URL),
			log.IntField("response_status", resp.StatusCode),
		)
		return nil
	}
	return fmt.Errorf("forgejo at %s answered %d to /api/v1/version", flag.URL, resp.StatusCode)
}
