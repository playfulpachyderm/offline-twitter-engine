import urllib
import urllib.parse as parse
import json

x = "https://twitter.com/i/api/graphql/3_7xfjmh897x8h_n6QBqTA/Followers?variables=%7B%22userId%22%3A%221488963321701171204%22%2C%22count%22%3A20%2C%22includePromotedContent%22%3Afalse%7D&features=%7B%22responsive_web_graphql_exclude_directive_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22creator_subscriptions_tweet_preview_api_enabled%22%3Atrue%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22responsive_web_graphql_skip_user_profile_image_extensions_enabled%22%3Afalse%2C%22c9s_tweet_anatomy_moderator_badge_enabled%22%3Atrue%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22view_counts_everywhere_api_enabled%22%3Atrue%2C%22longform_notetweets_consumption_enabled%22%3Atrue%2C%22responsive_web_twitter_article_tweet_consumption_enabled%22%3Afalse%2C%22tweet_awards_web_tipping_enabled%22%3Afalse%2C%22freedom_of_speech_not_reach_fetch_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Atrue%2C%22rweb_video_timestamps_enabled%22%3Atrue%2C%22longform_notetweets_rich_text_read_enabled%22%3Atrue%2C%22longform_notetweets_inline_media_enabled%22%3Atrue%2C%22responsive_web_media_download_video_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Afalse%7D"
parsed_url = parse.urlparse(x)

base_url = parsed_url._replace(query="").geturl()

gql_vars = json.loads(parse.parse_qs(parsed_url.query)["variables"][0])
gql_feats = json.loads(parse.parse_qs(parsed_url.query)["features"][0])

def snake_to_camel(s):
	return "".join(x.capitalize() for x in s.split("_"))

print("BaseUrl: \"{}\",".format(base_url))
print("Variables: GraphqlVariables{")
for k, v in gql_vars.items():
	print("\t{}: {},".format(snake_to_camel(k), json.dumps(v)))
print("},")
print("Features: GraphqlFeatures{")
for k, v in gql_feats.items():
	print("\t{}: {},".format(snake_to_camel(k), json.dumps(v)))
print("},")
