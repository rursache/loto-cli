package client

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/rursache/loto-cli/config"
)

const (
	baseLotoURL   = "https://www.loto.ro"
	baseBileteURL = "https://bilete.loto.ro"
)

// Client is the HTTP client for interacting with loto.ro and bilete.loto.ro
type Client struct {
	HTTP      *http.Client
	Config    *config.Config
	cookieJar *cookiejar.Jar
}

// New creates a new Client with a cookie jar and the given config
func New(cfg *config.Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client := &Client{
		HTTP: &http.Client{
			Jar: jar,
		},
		Config:    cfg,
		cookieJar: jar,
	}

	return client, nil
}

// newRequest creates a new HTTP request with standard browser headers
func (c *Client) newRequest(method, rawURL string) (*http.Request, error) {
	req, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.Config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,ro-RO;q=0.8,ro;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("DNT", "1")
	req.Header.Set("Pragma", "no-cache")

	return req, nil
}

// doRequest executes a request and checks for geo-blocking
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusGone {
		resp.Body.Close()
		return nil, fmt.Errorf("loto.ro requires a Romanian IP address. Please connect from Romania or use a VPN")
	}

	return resp, nil
}

// SetCookies sets cookies on the client's jar for a given URL
func (c *Client) SetCookies(rawURL string, cookies []*http.Cookie) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	c.cookieJar.SetCookies(u, cookies)
	return nil
}

// GetCookies returns cookies from the jar for a given URL
func (c *Client) GetCookies(rawURL string) ([]*http.Cookie, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return c.cookieJar.Cookies(u), nil
}
