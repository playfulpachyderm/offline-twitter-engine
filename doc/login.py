"""
Utility script for testing twitter login workflow
"""

# pylint: disable=invalid-name

import requests


guest_token_response = requests.post("https://api.twitter.com/1.1/guest/activate.json", headers={"Authorization": "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"})
assert guest_token_response.status_code == 200, f"HTTP Response code {guest_token_response.status_code}"
guest_token = guest_token_response.json()["guest_token"]


headers = {
    "authorization":'Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA',  # pylint: disable=line-too-long
    'content-type':'application/json',
    "x-guest-token":guest_token,
}
login_curl = "https://twitter.com/i/api/1.1/onboarding/task.json"

username = "offline_twatter"
password = "S1pKIW#eRT016iA@OFcK"

response = requests.post(login_curl, headers=headers, params={"flow_name": "login"})
assert response.status_code == 200, f"HTTP Response code {response.status_code}"
flow_token = response.json()['flow_token']

second_request_data = {
    "flow_token": flow_token,
    "subtask_inputs": [
        {
            "subtask_id": "LoginJsInstrumentationSubtask",
            "js_instrumentation": {
                "response": "{\"rf\":{\"a560cdc18ff70ce7662311eac0f2441dd3d3ed27c354f082f587e7a30d1a7d5f\":72,\"a8e890b5fec154e7af62d8f529fbec5942dfdd7ad41597245b71a3fbdc9a180d\":176,\"a3c24597ad4c862773c74b9194e675e96a7607708f6bbd0babcfdf8b109ed86d\":-161,\"af9847e2cd4e9a0ca23853da4b46bf00a2e801f98dc819ee0dd6ecc1032273fa\":-8},\"s\":\"hOai7h2KQi4RBGKSYLUhH0Y0fBm5KHIJgxD5AmNKtwP7N8gpVuAqP8o9n2FpCnNeR1d6XbB0QWkGAHiXkKao5PhaeXEZgPJU1neLcVgTnGuFzpjDnGutCUgYaxNiwUPfDX0eQkgr_q7GWmbB7yyYPt32dqSd5yt-KCpSt7MOG4aFmGf11xWE4MTpXfkefbnX4CwZeEFKQQYzJptOvmUWa7qI0A69BSOs7HZ_4Wry2TwB9k03Q_S-MDZAZ3yB_L7WoosVVb1e84YWgaLWWzqhz4C77jDy6isT8EKSWKWnVctsIcaqM_wMV8AiYa5lr0_WkN5TwK9h0vDOTS1obOZuhAAAAYTZan_3\"}",  # pylint: disable=line-too-long
                "link": "next_link"
            }
        }
    ]
}

response2 = requests.post(login_curl, headers=headers, json=second_request_data)
assert response2.status_code == 200, f"HTTP Response code {response2.status_code}"
flow_token = response2.json()["flow_token"]


third_request_data = {
    "flow_token": flow_token,
    "subtask_inputs": [
        {
            "subtask_id": "LoginEnterUserIdentifierSSO",
            "settings_list": {
                "setting_responses": [
                    {
                        "key": "user_identifier",
                        "response_data": {
                            "text_data": {
                                "result": username,
                            }
                        }
                    }
                ],
                "link": "next_link"
            }
        }
    ]
}

response3 = requests.post(login_curl, headers=headers, json=third_request_data)
assert response3.status_code == 200, f"HTTP Response code {response3.status_code}"
flow_token = response3.json()["flow_token"]

fourth_request_data = {
    "flow_token": flow_token,
    "subtask_inputs": [
        {
            "subtask_id": "LoginEnterPassword",
            "enter_password": {
                "password": password,
                "link": "next_link"
            }
        }
    ]
}

response4 = requests.post(login_curl, headers = headers, json=fourth_request_data)
assert response4.status_code == 200, f"HTTP Response code {response4.status_code}"
flow_token = response4.json()["flow_token"]

fifth_request_data = {
    "flow_token": flow_token,
    "subtask_inputs": [
        {
            "subtask_id": "AccountDuplicationCheck",
            "check_logged_in_account": {
                "link": "AccountDuplicationCheck_false"
            }
        }
    ]
}

response5 = requests.post(login_curl, headers = headers, json=fifth_request_data)
assert response5.status_code == 200, f"HTTP Response code {response5.status_code}"
flow_token = response5.json()["flow_token"]
cookie = response5.headers["set-cookie"]
cookie_dict = response5.cookies
print("cookie:", cookie)
print("cookie_dict:", cookie_dict)

print("csrf:", cookie_dict["ct0"])


likes_headers = {
    'authorization': 'Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA',
    'content-type': 'application/json',
    'x-csrf-token': cookie_dict["ct0"],
}

likes_url = 'https://twitter.com/i/api/graphql/2Z6LYO4UTM4BnWjaNCod6g/Likes?variables=%7B%22userId%22%3A%221458284524761075714%22%2C%22count%22%3A20%2C%22includePromotedContent%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withClientEventToken%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Atrue%7D&features=%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D'

likes_response = requests.get(likes_url, headers=likes_headers, cookies=cookie_dict)
assert likes_response.status_code == 200, f"HTTP Response code {likes_response.status_code}: {likes_response.text}"


print(likes_response)
print(likes_response.json())
