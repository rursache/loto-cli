package client

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rursache/loto-cli/config"
)

const cookiesFileName = "cookies.json"

// savedCookie is a serializable representation of an http.Cookie
type savedCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires"`
	Secure   bool      `json:"secure"`
	HttpOnly bool      `json:"http_only"`
}

// getCookiesPath returns the path to the cookies file
func getCookiesPath() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, cookiesFileName), nil
}

// LoadCookies reads saved cookies from disk and applies them to the client
func (c *Client) LoadCookies() error {
	cookiesPath, err := getCookiesPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(cookiesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no saved cookies, not an error
		}
		return err
	}

	var saved []savedCookie
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil // corrupted file, ignore
	}

	var cookies []*http.Cookie
	for _, sc := range saved {
		if sc.Expires.Before(time.Now()) {
			continue // skip expired cookies
		}
		cookies = append(cookies, &http.Cookie{
			Name:     sc.Name,
			Value:    sc.Value,
			Domain:   sc.Domain,
			Path:     sc.Path,
			Expires:  sc.Expires,
			Secure:   sc.Secure,
			HttpOnly: sc.HttpOnly,
		})
	}

	if len(cookies) > 0 {
		return c.SetCookies(baseBileteURL, cookies)
	}

	return nil
}

// SaveCookies writes the current session cookies to disk
func (c *Client) SaveCookies() error {
	cookies, err := c.GetCookies(baseBileteURL)
	if err != nil {
		return err
	}

	var saved []savedCookie
	for _, cookie := range cookies {
		saved = append(saved, savedCookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  cookie.Expires,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
		})
	}

	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}

	cookiesPath, err := getCookiesPath()
	if err != nil {
		return err
	}

	return os.WriteFile(cookiesPath, data, 0600)
}

// ClearCookies deletes the saved cookies file
func (c *Client) ClearCookies() error {
	cookiesPath, err := getCookiesPath()
	if err != nil {
		return err
	}
	if err := os.Remove(cookiesPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
