package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

const BEARER_TOKEN string = "AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"

type GuestTokenResponse struct {
	Token       string `json:"guest_token"`
	RefreshedAt time.Time
}

var guestToken GuestTokenResponse

func GetGuestTokenWithRetries(n int, sleep time.Duration) (ret string, err error) {
	for i := 0; i < n; i++ {
		ret, err = GetGuestToken()
		if err == nil {
			return
		}
		log.Printf("Failed to get guest token: %s\nRetrying...", err.Error())
		time.Sleep(sleep)
	}
	return
}

func GetGuestToken() (string, error) {
	// Guest token is still valid; no need for new one
	if time.Since(guestToken.RefreshedAt).Hours() < 1 {
		return guestToken.Token, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "https://api.twitter.com/1.1/guest/activate.json", nil)
	if err != nil {
		return "", fmt.Errorf("Error initializing HTTP request:\n  %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+BEARER_TOKEN)

	resp, err := client.Do(req)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.Err == "server misbehaving" && dnsErr.Temporary() {
			return "", ErrNoInternet
		}

		return "", fmt.Errorf("Error executing HTTP request:\n  %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		return "", fmt.Errorf("HTTP %s: %s", resp.Status, content)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading HTTP response body:\n  %w", err)
	}

	err = json.Unmarshal(body, &guestToken)
	if err != nil {
		return "", fmt.Errorf("Error parsing API response:\n  %w", err)
	}

	guestToken.RefreshedAt = time.Now()
	return guestToken.Token, nil
}
