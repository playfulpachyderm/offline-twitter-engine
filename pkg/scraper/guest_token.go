package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GuestTokenResponse struct {
	Token       string `json:"guest_token"`
	RefreshedAt time.Time
}

var guestToken GuestTokenResponse

func GetGuestToken() (string, error) {
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
