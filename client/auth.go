package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	loginURL           = baseBileteURL + "/login"
	authCheckURL       = baseBileteURL + "/history/ticket?page_no=1"
	authenticatedTitle = "Biletele Mele"
)

// Login authenticates with bilete.loto.ro using saved cookies or fresh credentials.
// It first attempts to restore a previous session from saved cookies.
// If no valid session exists, it performs a full login using the configured email and password.
func (c *Client) Login() error {
	// Try to restore session from saved cookies
	if err := c.LoadCookies(); err == nil {
		if c.IsAuthenticated() {
			return nil
		}
	}

	// Step 1: GET the login page to obtain the CSRF token
	csrfToken, err := c.fetchCSRFToken()
	if err != nil {
		return fmt.Errorf("failed to fetch CSRF token: %w", err)
	}

	// Step 2: POST credentials
	if err := c.postLogin(csrfToken); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Verify authentication succeeded
	if !c.IsAuthenticated() {
		return fmt.Errorf("login failed: invalid credentials or unexpected response")
	}

	// Persist the session cookies
	if err := c.SaveCookies(); err != nil {
		return fmt.Errorf("failed to save cookies: %w", err)
	}

	return nil
}

// IsAuthenticated checks whether the current session is still valid by requesting
// the ticket history page and inspecting the page title.
// Returns true if the session is authenticated, false otherwise.
func (c *Client) IsAuthenticated() bool {
	req, err := c.newRequest(http.MethodGet, authCheckURL)
	if err != nil {
		return false
	}

	// Don't follow redirects - a 302 to /login means session expired
	c.HTTP.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer func() {
		c.HTTP.CheckRedirect = nil
	}()

	resp, err := c.doRequest(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// A redirect means the session is expired
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		return false
	}

	if resp.StatusCode != http.StatusOK {
		return false
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false
	}

	title := doc.Find("title").First().Text()
	return strings.Contains(title, authenticatedTitle)
}

// fetchCSRFToken performs a GET request to the login page and extracts the CSRF token
// from the <meta name="csrf-token"> tag.
func (c *Client) fetchCSRFToken() (string, error) {
	req, err := c.newRequest(http.MethodGet, loginURL)
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from login page", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse login page: %w", err)
	}

	token, exists := doc.Find(`meta[name="csrf-token"]`).Attr("content")
	if !exists || token == "" {
		return "", fmt.Errorf("CSRF token not found on login page")
	}

	return token, nil
}

// postLogin submits the login form with the CSRF token and user credentials.
func (c *Client) postLogin(csrfToken string) error {
	formData := url.Values{
		"_token":   {csrfToken},
		"email":    {c.Config.Email},
		"password": {c.Config.Password},
	}

	req, err := c.newRequest(http.MethodPost, loginURL)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", loginURL)
	req.Body = io.NopCloser(strings.NewReader(formData.Encode()))
	req.ContentLength = int64(len(formData.Encode()))

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Laravel redirects on successful login (302)
	// The cookie jar automatically captures the new session cookies
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("unexpected status %d after login POST", resp.StatusCode)
	}

	return nil
}
