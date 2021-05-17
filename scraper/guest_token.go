package scraper

import "fmt"
import "time"

import "io/ioutil"
import "net/http"
import "encoding/json"

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
		return "", err
	}
	req.Header.Set("Authorization", "Bearer " + BEARER_TOKEN)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %s: %s", resp.Status, content)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &guestToken)
	if err != nil {
		return "", err
	}

	guestToken.RefreshedAt = time.Now()
	return guestToken.Token, nil
}
