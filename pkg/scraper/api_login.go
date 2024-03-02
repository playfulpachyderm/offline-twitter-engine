package scraper

import (
	"encoding/json"
	"fmt"
	"strings"
)

const LOGIN_URL = "https://twitter.com/i/api/1.1/onboarding/task.json"

type flow_result struct {
	FlowToken string `json:"flow_token"`
	Subtasks  []struct {
		SubtaskID string `json:"subtask_id"`

		// Login success
		OpenAccount struct {
			User struct {
				ID         int `json:"id_str,string"`
				Name       string
				ScreenName string `json:"screen_name"`
			}
		} `json:"open_account"`

		// Phone verification challenge
		EnterText struct {
			PrimaryText struct {
				Text string `json:"text"`
			} `json:"primary_text"`
			SecondaryText struct {
				Text string `json:"text"`
			} `json:"secondary_text"`
		} `json:"enter_text"`
	} `json:"subtasks"`
}

type ChallengeResult struct {
	FlowToken     string
	SubtaskID     string
	Data          map[string]string
	PrimaryText   string
	SecondaryText string
}

func (api *API) do_login_task(flow_token string, task_id string, data map[string]string) (result flow_result) {
	var body string
	switch task_id {
	// Regular login flow
	case "LoginJsInstrumentationSubtask":
		body = fmt.Sprintf(
			`{"flow_token":"%s", "subtask_inputs": [{"subtask_id": "LoginJsInstrumentationSubtask", "js_instrumentation": {"response": "{\"rf\":{\"a560cdc18ff70ce7662311eac0f2441dd3d3ed27c354f082f587e7a30d1a7d5f\":72,\"a8e890b5fec154e7af62d8f529fbec5942dfdd7ad41597245b71a3fbdc9a180d\":176,\"a3c24597ad4c862773c74b9194e675e96a7607708f6bbd0babcfdf8b109ed86d\":-161,\"af9847e2cd4e9a0ca23853da4b46bf00a2e801f98dc819ee0dd6ecc1032273fa\":-8},\"s\":\"hOai7h2KQi4RBGKSYLUhH0Y0fBm5KHIJgxD5AmNKtwP7N8gpVuAqP8o9n2FpCnNeR1d6XbB0QWkGAHiXkKao5PhaeXEZgPJU1neLcVgTnGuFzpjDnGutCUgYaxNiwUPfDX0eQkgr_q7GWmbB7yyYPt32dqSd5yt-KCpSt7MOG4aFmGf11xWE4MTpXfkefbnX4CwZeEFKQQYzJptOvmUWa7qI0A69BSOs7HZ_4Wry2TwB9k03Q_S-MDZAZ3yB_L7WoosVVb1e84YWgaLWWzqhz4C77jDy6isT8EKSWKWnVctsIcaqM_wMV8AiYa5lr0_WkN5TwK9h0vDOTS1obOZuhAAAAYTZan_3\"}", "link": "next_link"}}]}`, //nolint:lll  // json body
			flow_token)
	case "LoginEnterUserIdentifierSSO":
		username_json, is_ok := data["username_json"]
		if !is_ok {
			panic("No username provided")
		}
		body = fmt.Sprintf(
			`{"flow_token":"%s","subtask_inputs":[{"subtask_id":"LoginEnterUserIdentifierSSO","settings_list":{"setting_responses":[{"key":"user_identifier","response_data":{"text_data":{"result":`+username_json+`}}}],"link":"next_link"}}]}`, //nolint:lll  // json body
			flow_token)
	case "LoginEnterPassword":
		password_json, is_ok := data["password_json"]
		if !is_ok {
			panic("No password provided")
		}
		body = fmt.Sprintf(
			`{"flow_token":"%s","subtask_inputs":[{"subtask_id":"LoginEnterPassword","enter_password":{"password":`+password_json+`,"link":"next_link"}}]}`, //nolint:lll  // json body
			flow_token)
	case "AccountDuplicationCheck":
		body = fmt.Sprintf(
			`{"flow_token":"%s","subtask_inputs":[{"subtask_id":"AccountDuplicationCheck","check_logged_in_account":{"link":"AccountDuplicationCheck_false"}}]}`, //nolint:lll  // json body
			flow_token)

	// Challenge flows
	case "LoginAcid":
		body = fmt.Sprintf(
			`{"flow_token":"%s","subtask_inputs":[{"subtask_id":"LoginAcid","enter_text":{"text":"%s","link":"next_link"}}]}`,
			flow_token,
			data["phone"])
	case "LoginEnterAlternateIdentifierSubtask":
		body = fmt.Sprintf(
			`{"flow_token":"%s","subtask_inputs":[{"subtask_id":"LoginEnterAlternateIdentifierSubtask","enter_text":{"text":"%s","link":"next_link"}}]}`,
			flow_token,
			data["phone"])

	default:
		panic("Unknown task_id: " + task_id)
	}

	err := api.do_http_POST(LOGIN_URL, body, &result)
	if err != nil {
		fmt.Printf("api.Client.Jar: %#v\n", api.Client.Jar)
		panic(err)
	}
	return
}

// Conducts the "login flow".
//
// Logging in is implemented as a series of subtasks in which username and password are submitted in
// separate requests ("tasks").  Other possible tasks include answering challenges like "verify your
// phone number", etc.
//
// To log in, we do "flow tasks" in sequence until the result is "LoginSuccessSubtask"
func (api *API) LogIn(username string, password string) *ChallengeResult {
	// Format username and password safely as JSON (escape quotes, etc)
	username_json, err := json.Marshal(username)
	if err != nil {
		panic(err)
	}
	password_json, err := json.Marshal(password)
	if err != nil {
		panic(err)
	}

	data := map[string]string{
		"username_json": string(username_json),
		"password_json": string(password_json),
	}

	// Begin flow
	var result flow_result
	err = api.do_http_POST(LOGIN_URL+"?flow_name=login", "", &result)
	if err != nil {
		panic(err)
	}

	// Continue login flow
	return api.continue_login_flow(result, data)
}

// Continue flow until finished or challenged
// This helper function lets you re-enter the login flow after a challenge disrupts it
func (api *API) continue_login_flow(result flow_result, data map[string]string) *ChallengeResult {
	for result.Subtasks[0].SubtaskID != "LoginSuccessSubtask" {
		if result.FlowToken == "" { // Sanity check
			panic("No flow token.")
		}
		result = api.do_login_task(result.FlowToken, result.Subtasks[0].SubtaskID, data)

		// Check for challenges
		if result.Subtasks[0].EnterText.PrimaryText.Text != "" {
			// Challenge issued
			return &ChallengeResult{
				FlowToken:     result.FlowToken,
				Data:          data,
				SubtaskID:     result.Subtasks[0].SubtaskID,
				PrimaryText:   result.Subtasks[0].EnterText.PrimaryText.Text,
				SecondaryText: result.Subtasks[0].EnterText.SecondaryText.Text,
			}
		}
	}

	// Login successful
	api.UserID = UserID(result.Subtasks[0].OpenAccount.User.ID)
	api.UserHandle = UserHandle(result.Subtasks[0].OpenAccount.User.ScreenName)
	api.update_csrf_token()
	api.IsAuthenticated = true

	return nil
}

func (api *API) LoginVerifyPhone(challenge ChallengeResult, phone_num string) {
	data := challenge.Data
	data["phone"] = strings.TrimSpace(phone_num)
	result := api.do_login_task(challenge.FlowToken, challenge.SubtaskID, data)
	api.continue_login_flow(result, data)
}
