"""
Utility script for testing twitter login workflow
"""

# pylint: disable=invalid-name

import requests
import os

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
password = os.environ.get("OFFLINE_TWATTER_PASSWD")
if not password:
    print("No password provided!  Please set OFFLINE_TWATTER_PASSWD environment variable and try again.")

response = requests.post(login_curl, headers=headers, params={"flow_name": "login"})
assert response.status_code == 200, f"HTTP Response code {response.status_code}{response.json()}"
flow_token = response.json()['flow_token']

second_request_data = {
    "flow_token": flow_token,
    "subtask_inputs": [
        {
            "subtask_id": "LoginJsInstrumentationSubtask",
            "js_instrumentation": {
                "response": "{\"rf\":{\"cbd6bdbb6add3e20bbda71c0cd9f43a2f00533d3d6549a7ccb89c4c06901dce2\":1,\"ae977661f2a886f4ffacdd57540338e2b7aef11c4113806eceae3d22055aa8e7\":-149,\"dd84b8c06765cf9bf1fb433be2f59155facb2873a28c957297de3a1e993494fb\":-143,\"ab1534f3890a2597699a4594ad9f54e8432fed1291abfd96c892620f49baa907\":-184},\"s\":\"XpY8TxYmi24ixccqjW68xUdUsWJhmMJgDdf0oEA99ufun62H3AJRHJT1f2-gsxrCa289ZjcvOrXC7C5miGlWlofiwpF-nK7bIH1jCW_Jp6_9NiXaGH151Kt3ChCfqwYNv7gROBkyXOVMXxV2I8WZ_WF1Da1r2DlMVK9DosRhZHGJUhYDtJRhn65gi5Xd73z8MjOWODXsDMxm_urFU5bY68Arf2D1oUJcZ70jhjMYV0yU09249xWiXIs81n0i_44dYqWV2tuYFkdU7kSazjYz4VGZ4P70l1k_MwUuI6dbueK9-R1RBRszGNle0kVZmpJNmVocc3j3TMxOxwso29_cEwAAAYUm0EPM\"}",
                "link": "next_link"
            }
        }
    ]
}

response2 = requests.post(login_curl, headers=headers, json=second_request_data, cookies=response.cookies)
assert response2.status_code == 200, f"HTTP Response code {response2.status_code}, {response2.json()}"
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

response3 = requests.post(login_curl, headers=headers, json=third_request_data, cookies=response.cookies)
assert response3.status_code == 200, f"HTTP Response code {response3.status_code}, {response3.json()}"
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

response4 = requests.post(login_curl, headers = headers, json=fourth_request_data, cookies=response.cookies)
assert response4.status_code == 200, f"HTTP Response code {response4.status_code}: {response4.json()}"
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

response5 = requests.post(login_curl, headers = headers, json=fifth_request_data, cookies=response.cookies)
assert response5.status_code == 200, f"HTTP Response code {response5.status_code}: {response5.json()}"
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

feed_url = "https://twitter.com/i/api/graphql/CwLU7qTfeu0doqhSr6tW4A/UserTweetsAndReplies?variables=%7B%22userId%22%3A%221458284524761075714%22%2C%22count%22%3A40%2C%22includePromotedContent%22%3Afalse%2C%22withCommunity%22%3Atrue%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withBirdwatchPivots%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Afalse%2C%22__fs_interactive_text%22%3Afalse%2C%22__fs_dont_mention_me_view_api_enabled%22%3Afalse%7D"

#likes_response = requests.get(feed_url, headers=likes_headers, cookies=cookie_dict)
#assert likes_response.status_code == 200, f"HTTP Response code {likes_response.status_code}: {likes_response.text}"


#print(likes_response)
#print(likes_response.json())

dm_url = "https://twitter.com/i/api/1.1/dm/inbox_initial_state.json?nsfw_filtering_enabled=false&filter_low_quality=true&include_quality=all&include_profile_interstitial_type=1&include_blocking=1&include_blocked_by=1&include_followed_by=1&include_want_retweets=1&include_mute_edge=1&include_can_dm=1&include_can_media_tag=1&include_ext_has_nft_avatar=1&include_ext_is_blue_verified=1&include_ext_verified_type=1&include_ext_profile_image_shape=1&skip_status=1&dm_secret_conversations_enabled=false&krs_registration_enabled=true&cards_platform=Web-12&include_cards=1&include_ext_alt_text=true&include_ext_limited_action_results=false&include_quote_count=true&include_reply_count=1&tweet_mode=extended&include_ext_views=true&dm_users=true&include_groups=true&include_inbox_timelines=true&include_ext_media_color=true&supports_reactions=true&include_ext_edit_control=true&ext=mediaColor%2CaltText%2CmediaStats%2ChighlightedLabel%2ChasNftAvatar%2CvoiceInfo%2CbirdwatchPivot%2Cenrichments%2CsuperFollowMetadata%2CunmentionInfo%2CeditControl%2Cvibe"

dm_response = requests.get(dm_url, headers=likes_headers.copy(), cookies=cookie_dict)
assert dm_response.status_code == 200, f"HTTP Response code {dm_response.status_code}: {dm_response.text}"

print(dm_response)
print(dm_response.json())
