import requests, json

headers = {
    "authorization":'Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA',
    "x-guest-token":'1599119990820671490'
}
login_curl = "https://twitter.com/i/api/1.1/onboarding/task.json"

password = "S1pKIW#eRT016iA@OFcK"

response = requests.post("https://twitter.com/i/api/1.1/onboarding/task.json?flow_name=login", headers= headers)

#print("first request:",response)
#print(response.json())
print("first request cookie:",response.headers["set-cookie"])

first_request_flow_token = response.json()['flow_token']
#print(first_request_flow_token)

second_request_headers = {
#   'authority':'twitter.com',
#   'accept':'*/*',
#   'accept-language':'en-US,en;q=0.5',
  'authorization':'Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA',
  'content-type':'application/json',
#   'cookie':'guest_id_marketing=v1%3A167009488980761169; guest_id_ads=v1%3A167009488980761169; personalization_id="v1_Ct7NiaYdABX/4edgcqPmrg=="; guest_id=v1%3A167009488980761169; ct0=589750250f2314ad17bcd8354182d31c; gt=1599119990820671490; external_referer=padhuUp37zhzf%2BzW9lbDSEb6StpkXI7fDirGjZNZuxk%3D|0|8e8t2xd8A2w%3D; att=1-O9RWSFJgGzfOcY8fn867lojBkfbfamGEOLarnb4G; _twitter_sess=BAh7CSIKZmxhc2hJQzonQWN0aW9uQ29udHJvbGxlcjo6Rmxhc2g6OkZsYXNo%250ASGFzaHsABjoKQHVzZWR7ADoPY3JlYXRlZF9hdGwrCO5%252FatmEAToMY3NyZl9p%250AZCIlYTYwZjcxNzMzNzE1MzJhNjJkYjZjZjg4NWNjZjBiMTg6B2lkIiVkNjBk%250AMTQyZTQ0NmYyYWFjMjg0OGI4MzlkYzRlMWI5YQ%253D%253D--c73a9d9452a7cc26692dda28ede591a669d98bb3',
#   'origin':'https://twitter.com',
#   'referer':'https://twitter.com/i/flow/login',
#   'sec-ch-ua':'"Chromium";v="106", "Brave Browser";v="106", "Not;A=Brand";v="99"',
#   'sec-ch-ua-mobile':'?0',
#   'sec-ch-ua-platform':'"macOS"',
#   'sec-fetch-dest':'empty',
#   'sec-fetch-mode':'cors',
#   'sec-fetch-site':'same-origin',
#   'sec-gpc':'1',
#   'user-agent':'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36',
#   'x-csrf-token':'589750250f2314ad17bcd8354182d31c',
  'x-guest-token':'1599119990820671490',
#   'x-twitter-active-user':'yes',
#   'x-twitter-client-language':'en',
}

second_request_data = {
    "flow_token": first_request_flow_token,
    "subtask_inputs": [
        {
            "subtask_id": "LoginJsInstrumentationSubtask",
            "js_instrumentation": {
                "response": "{\"rf\":{\"a560cdc18ff70ce7662311eac0f2441dd3d3ed27c354f082f587e7a30d1a7d5f\":72,\"a8e890b5fec154e7af62d8f529fbec5942dfdd7ad41597245b71a3fbdc9a180d\":176,\"a3c24597ad4c862773c74b9194e675e96a7607708f6bbd0babcfdf8b109ed86d\":-161,\"af9847e2cd4e9a0ca23853da4b46bf00a2e801f98dc819ee0dd6ecc1032273fa\":-8},\"s\":\"hOai7h2KQi4RBGKSYLUhH0Y0fBm5KHIJgxD5AmNKtwP7N8gpVuAqP8o9n2FpCnNeR1d6XbB0QWkGAHiXkKao5PhaeXEZgPJU1neLcVgTnGuFzpjDnGutCUgYaxNiwUPfDX0eQkgr_q7GWmbB7yyYPt32dqSd5yt-KCpSt7MOG4aFmGf11xWE4MTpXfkefbnX4CwZeEFKQQYzJptOvmUWa7qI0A69BSOs7HZ_4Wry2TwB9k03Q_S-MDZAZ3yB_L7WoosVVb1e84YWgaLWWzqhz4C77jDy6isT8EKSWKWnVctsIcaqM_wMV8AiYa5lr0_WkN5TwK9h0vDOTS1obOZuhAAAAYTZan_3\"}",
                "link": "next_link"
            }
        }
    ]
}

second_response_post = requests.post(login_curl, headers = second_request_headers, data = json.dumps(second_request_data))
# print("second request:", second_response_post)
# print("second requiest json:",second_response_post.json())
print("second request cookie:",second_response_post.headers["set-cookie"])

second_request_flow_token = second_response_post.json()["flow_token"]


third_request_headers = {
#   'authority':'twitter.com',
#   'accept':'*/*',
#   'accept-language':'en-US,en;q=0.5',
  'authorization':'Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA',
  'content-type':'application/json',
#   'cookie':'guest_id_marketing=v1%3A167009488980761169; guest_id_ads=v1%3A167009488980761169; personalization_id="v1_Ct7NiaYdABX/4edgcqPmrg=="; guest_id=v1%3A167009488980761169; ct0=589750250f2314ad17bcd8354182d31c; gt=1599119990820671490; external_referer=padhuUp37zhzf%2BzW9lbDSEb6StpkXI7fDirGjZNZuxk%3D|0|8e8t2xd8A2w%3D; att=1-O9RWSFJgGzfOcY8fn867lojBkfbfamGEOLarnb4G; _twitter_sess=BAh7CSIKZmxhc2hJQzonQWN0aW9uQ29udHJvbGxlcjo6Rmxhc2g6OkZsYXNo%250ASGFzaHsABjoKQHVzZWR7ADoPY3JlYXRlZF9hdGwrCO5%252FatmEAToMY3NyZl9p%250AZCIlYTYwZjcxNzMzNzE1MzJhNjJkYjZjZjg4NWNjZjBiMTg6B2lkIiVkNjBk%250AMTQyZTQ0NmYyYWFjMjg0OGI4MzlkYzRlMWI5YQ%253D%253D--c73a9d9452a7cc26692dda28ede591a669d98bb3',
#   'origin':'https://twitter.com',
#   'referer':'https://twitter.com/i/flow/login',
#   'sec-ch-ua':'"Chromium";v="106", "Brave Browser";v="106", "Not;A=Brand";v="99"',
#   'sec-ch-ua-mobile':'?0',
#   'sec-ch-ua-platform':'"macOS"',
#   'sec-fetch-dest':'empty',
#   'sec-fetch-mode':'cors',
#   'sec-fetch-site':'same-origin',
#   'sec-gpc':'1',
#   'user-agent':'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36',
#   'x-csrf-token':'589750250f2314ad17bcd8354182d31c',
  'x-guest-token':'1599119990820671490',
#   'x-twitter-active-user':'yes',
#   'x-twitter-client-language':'en',
}

third_request_data = {
    "flow_token": second_request_flow_token,
    "subtask_inputs": [
        {
            "subtask_id": "LoginEnterUserIdentifierSSO",
            "settings_list": {
                "setting_responses": [
                    {
                        "key": "user_identifier",
                        "response_data": {
                            "text_data": {
                                "result": "offline_twatter"
                            }
                        }
                    }
                ],
                "link": "next_link"
            }
        }
    ]
}

third_response_post = requests.post(login_curl, headers = third_request_headers, data = json.dumps(third_request_data))
# print("third request:", third_response_post)
# print("third requiest json:",third_response_post.json())
print("third request cookie:",third_response_post.headers["set-cookie"])

third_request_flow_token = third_response_post.json()["flow_token"]

fourth_request_headers = {
  'authority':'twitter.com',
  'accept':'*/*',
  'accept-language':'en-US,en;q=0.5',
  'authorization':'Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA',
  'content-type':'application/json',
  'cookie':'guest_id_marketing=v1%3A167009488980761169; guest_id_ads=v1%3A167009488980761169; personalization_id="v1_Ct7NiaYdABX/4edgcqPmrg=="; guest_id=v1%3A167009488980761169; ct0=589750250f2314ad17bcd8354182d31c; gt=1599119990820671490; external_referer=padhuUp37zhzf%2BzW9lbDSEb6StpkXI7fDirGjZNZuxk%3D|0|8e8t2xd8A2w%3D; att=1-O9RWSFJgGzfOcY8fn867lojBkfbfamGEOLarnb4G; _twitter_sess=BAh7CSIKZmxhc2hJQzonQWN0aW9uQ29udHJvbGxlcjo6Rmxhc2g6OkZsYXNo%250ASGFzaHsABjoKQHVzZWR7ADoPY3JlYXRlZF9hdGwrCO5%252FatmEAToMY3NyZl9p%250AZCIlYTYwZjcxNzMzNzE1MzJhNjJkYjZjZjg4NWNjZjBiMTg6B2lkIiVkNjBk%250AMTQyZTQ0NmYyYWFjMjg0OGI4MzlkYzRlMWI5YQ%253D%253D--c73a9d9452a7cc26692dda28ede591a669d98bb3',
  'origin':'https://twitter.com',
  'referer':'https://twitter.com/i/flow/login',
  'sec-ch-ua':'"Chromium";v="106", "Brave Browser";v="106", "Not;A=Brand";v="99"',
  'sec-ch-ua-mobile':'?0',
  'sec-ch-ua-platform':'"macOS"',
  'sec-fetch-dest':'empty',
  'sec-fetch-mode':'cors',
  'sec-fetch-site':'same-origin',
  'sec-gpc':'1',
  'user-agent':'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36',
  'x-csrf-token':'589750250f2314ad17bcd8354182d31c',
  'x-guest-token':'1599119990820671490',
  'x-twitter-active-user':'yes',
  'x-twitter-client-language':'en',
}
fourth_request_data = {
    "flow_token": third_request_flow_token,
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

fourth_response_post = requests.post(login_curl, headers = fourth_request_headers, data = json.dumps(fourth_request_data))
# print("fourth request:", fourth_response_post)
# print("fourth requiest json:",fourth_response_post.json())
print("fourth request cookie:",fourth_response_post.headers["set-cookie"])

fourth_request_flow_token = fourth_response_post.json()["flow_token"]

fifth_request_headers = {
  'authority':'twitter.com',
  'accept':'*/*',
  'accept-language':'en-US,en;q=0.5',
  'authorization':'Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA',
  'content-type':'application/json',
  'cookie':'guest_id_marketing=v1%3A167009488980761169; guest_id_ads=v1%3A167009488980761169; personalization_id="v1_Ct7NiaYdABX/4edgcqPmrg=="; guest_id=v1%3A167009488980761169; ct0=589750250f2314ad17bcd8354182d31c; gt=1599119990820671490; external_referer=padhuUp37zhzf%2BzW9lbDSEb6StpkXI7fDirGjZNZuxk%3D|0|8e8t2xd8A2w%3D; att=1-O9RWSFJgGzfOcY8fn867lojBkfbfamGEOLarnb4G; _twitter_sess=BAh7CSIKZmxhc2hJQzonQWN0aW9uQ29udHJvbGxlcjo6Rmxhc2g6OkZsYXNo%250ASGFzaHsABjoKQHVzZWR7ADoPY3JlYXRlZF9hdGwrCO5%252FatmEAToMY3NyZl9p%250AZCIlYTYwZjcxNzMzNzE1MzJhNjJkYjZjZjg4NWNjZjBiMTg6B2lkIiVkNjBk%250AMTQyZTQ0NmYyYWFjMjg0OGI4MzlkYzRlMWI5YQ%253D%253D--c73a9d9452a7cc26692dda28ede591a669d98bb3',
  'origin':'https://twitter.com',
  'referer':'https://twitter.com/i/flow/login',
  'sec-ch-ua':'"Chromium";v="106", "Brave Browser";v="106", "Not;A=Brand";v="99"',
  'sec-ch-ua-mobile':'?0',
  'sec-ch-ua-platform':'"macOS"',
  'sec-fetch-dest':'empty',
  'sec-fetch-mode':'cors',
  'sec-fetch-site':'same-origin',
  'sec-gpc':'1',
  'user-agent':'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36',
  'x-csrf-token':'589750250f2314ad17bcd8354182d31c',
  'x-guest-token':'1599119990820671490',
  'x-twitter-active-user':'yes',
  'x-twitter-client-language':'en',
}

fifth_request_data = {
    "flow_token": fourth_request_flow_token,
    "subtask_inputs": [
        {
            "subtask_id": "AccountDuplicationCheck",
            "check_logged_in_account": {
                "link": "AccountDuplicationCheck_false"
            }
        }
    ]
}

fifth_response_post = requests.post(login_curl, headers = fifth_request_headers, data = json.dumps(fifth_request_data))
print("fifth request:", fifth_response_post)
print("fifth requiest json:",fifth_response_post.json())
cookie = fifth_response_post.headers["set-cookie"]
print("fifth request cookie:",fifth_response_post.headers["set-cookie"])

fifth_request_flow_token = fifth_response_post.json()["flow_token"]

print("fifth request flow_token:",fifth_request_flow_token)


likes_test_url = 'https://twitter.com/i/api/graphql/2Z6LYO4UTM4BnWjaNCod6g/Likes?variables=%7B%22userId%22%3A%221488963321701171204%22%2C%22count%22%3A20%2C%22includePromotedContent%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withClientEventToken%22%3Afalse%2C%22withBirdwatchNotes%22%3Afalse%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Atrue%7D&features=%7B%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22unified_cards_ad_metadata_container_dynamic_card_content_query_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D' \

likes_test_request = {
  'authority':'twitter.com',
  'accept':'*/*',
  'accept-language':'en-US,en;q=0.5',
  'authorization':'Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA',
  'content-type':'application/json',
  'cookie':'guest_id_marketing=v1%3A167009488980761169; guest_id_ads=v1%3A167009488980761169; personalization_id="v1_Ct7NiaYdABX/4edgcqPmrg=="; guest_id=v1%3A167009488980761169; gt=1599119990820671490; external_referer=padhuUp37zhzf%2BzW9lbDSEb6StpkXI7fDirGjZNZuxk%3D|0|8e8t2xd8A2w%3D; _twitter_sess=BAh7CSIKZmxhc2hJQzonQWN0aW9uQ29udHJvbGxlcjo6Rmxhc2g6OkZsYXNo%250ASGFzaHsABjoKQHVzZWR7ADoPY3JlYXRlZF9hdGwrCO5%252FatmEAToMY3NyZl9p%250AZCIlYTYwZjcxNzMzNzE1MzJhNjJkYjZjZjg4NWNjZjBiMTg6B2lkIiVkNjBk%250AMTQyZTQ0NmYyYWFjMjg0OGI4MzlkYzRlMWI5YQ%253D%253D--c73a9d9452a7cc26692dda28ede591a669d98bb3; kdt=mkM3Jke0Zh8DT1SI8k0u96lGaBpUBnIxVjpRzROq; auth_token=bfb59318bfaed4e46058c15f6d1ea8481f45f4b9; ct0=1657ba123e97ed7526e05048bf02ec94cc9d5853cac9c7ce1174b95fb7ed6a3f685a41c10b9155f35fee603529e2e833fa0ceace0f774c075dfeca5346b4333a0cfe7d0a9e0a3a61f63a3e746723fb51; att=1-TIlqaxCpfBErCey6ClocjZjuBXqLdQ3oZiFz4Lwm; twid=u%3D1488963321701171204',
  'referer':'https://twitter.com/Offline_Twatter/likes',
  'sec-ch-ua':'"Chromium";v="106", "Brave Browser";v="106", "Not;A=Brand";v="99"',
  'sec-ch-ua-mobile':'?0',
  'sec-ch-ua-platform':'"macOS"',
  'sec-fetch-dest':'empty',
  'sec-fetch-mode':'cors',
  'sec-fetch-site':'same-origin',
  'sec-gpc':'1',
  'user-agent':'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36',
  'x-csrf-token':'1657ba123e97ed7526e05048bf02ec94cc9d5853cac9c7ce1174b95fb7ed6a3f685a41c10b9155f35fee603529e2e833fa0ceace0f774c075dfeca5346b4333a0cfe7d0a9e0a3a61f63a3e746723fb51',
  'x-guest-token':'1599119990820671490',
  'x-twitter-active-user':'yes',
  'x-twitter-client-language':'en',

}