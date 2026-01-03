package enricher

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxRedirects = 10

func newSafeHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			return validateParsedURL(req.Context(), req.URL)
		},
	}
}

func validateFetchURL(ctx context.Context, rawURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if err := validateParsedURL(ctx, parsedURL); err != nil {
		return nil, err
	}
	return parsedURL, nil
}

func validateParsedURL(ctx context.Context, parsedURL *url.URL) error {
	if parsedURL == nil {
		return fmt.Errorf("invalid URL: empty")
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid URL: missing scheme or host")
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("invalid URL: unsupported scheme")
	}

	if parsedURL.User != nil {
		return fmt.Errorf("invalid URL: userinfo not allowed")
	}

	host := parsedURL.Hostname()
	if host == "" {
		return fmt.Errorf("invalid URL: missing host")
	}
	if isLocalhost(host) {
		return fmt.Errorf("invalid URL: host is not allowed")
	}

	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return fmt.Errorf("invalid URL: host resolves to private IP")
		}
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("invalid URL: failed to resolve host")
	}
	if len(addrs) == 0 {
		return fmt.Errorf("invalid URL: host has no addresses")
	}
	for _, addr := range addrs {
		if isPrivateIP(addr.IP) {
			return fmt.Errorf("invalid URL: host resolves to private IP")
		}
	}

	return nil
}

func isLocalhost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "localhost" || strings.HasSuffix(host, ".localhost")
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified()
}
