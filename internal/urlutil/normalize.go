package urlutil

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

var trackingParams = map[string]struct{}{
	"fbclid":          {},
	"gclid":           {},
	"mc_cid":          {},
	"mc_eid":          {},
	"ref":             {},
	"ref_src":         {},
	"utm_campaign":    {},
	"utm_content":     {},
	"utm_id":          {},
	"utm_medium":      {},
	"utm_source":      {},
	"utm_term":        {},
	"utm_reader":      {},
	"utm_name":        {},
	"utm_referrer":    {},
	"utm_social":      {},
	"utm_social_type": {},
}

// NormalizeURL normalizes a URL for duplicate detection.
func NormalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("empty url")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid url")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	host := strings.ToLower(parsed.Hostname())
	port := parsed.Port()
	if (parsed.Scheme == "http" && port == "80") || (parsed.Scheme == "https" && port == "443") {
		port = ""
	}
	if port != "" {
		parsed.Host = net.JoinHostPort(host, port)
	} else {
		parsed.Host = host
	}

	parsed.Fragment = ""

	if parsed.Path != "/" {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
		if parsed.Path == "" {
			parsed.Path = "/"
		}
	}
	parsed.RawPath = ""

	query := parsed.Query()
	for key := range query {
		if _, ok := trackingParams[strings.ToLower(key)]; ok {
			query.Del(key)
		}
	}
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}
