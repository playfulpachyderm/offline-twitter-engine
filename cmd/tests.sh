#!/bin/bash

set -e
set -x

PS4='+(${BASH_SOURCE}:${LINENO}): '

if [[ -z "$OFFLINE_TWATTER_PASSWD" ]]; then
    echo "OFFLINE_TWATTER_PASSWD not set!  Exiting."
    exit 1
fi
FAKE_VERSION="1.100.3489"
./compile.sh $FAKE_VERSION

test -e data && rm -r data

PATH=`pwd`:$PATH

test "$(tw --version)" = "v$FAKE_VERSION"

tw --help
test $? -eq 0

tw create_profile data
cd data

# Should only contain default profile image
test $(find profile_images | wc -l) = "2"
test -f profile_images/default_profile.png



# Print an error message in red before exiting if a test fails
trap 'echo -e "\033[31mTEST FAILURE.  Aborting\033[0m"' ERR

# Testing login
tw login offline_twatter "$OFFLINE_TWATTER_PASSWD"
test -f Offline_Twatter.session
test "$(jq .UserHandle Offline_Twatter.session)" = "\"Offline_Twatter\""
test "$(jq .IsAuthenticated Offline_Twatter.session)" = "true"
jq .CSRFToken Offline_Twatter.session | grep -P '"\w+"'

shopt -s expand_aliases
alias tw="tw --session Offline_Twatter"

# Fetch a user
initial_profile_images_count=$(find profile_images | wc -l)
tw fetch_user wrathofgnon
test "$(sqlite3 twitter.db "select handle from users")" = "wrathofgnon"
test $(sqlite3 twitter.db "select count(*) from users") = "1"
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle = 'wrathofgnon'") = "1"
test $(find profile_images | wc -l) = "$(($initial_profile_images_count + 2))"    # should have gotten 2 images
test -f profile_images/wrathofgnon_profile_fB-3BRin.jpg
test -f profile_images/wrathofgnon_banner_1503908468.jpg
tw fetch_user wrathofgnon                                      # try to double-download it
test $(sqlite3 twitter.db "select count(*) from users") = "1"  # shouldn't have added a new row


# # Fetch a tweet with images
tw fetch_tweet_only https://twitter.com/wrathofgnon/status/1503016316642689026
test $(sqlite3 twitter.db "select count(*) from tweets") = "1"
test "$(sqlite3 twitter.db "select text from tweets")" = "The I Am A Mayor Who Is Serious About Reducing My Town's Dependence on Fossil Fuels Starter Pack. Inquire within for more details."
test $(sqlite3 twitter.db "select count(*) from images") = "4"

# Download its images
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1503016316642689026 and is_downloaded = 0") = "4"
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1503016316642689026 and is_downloaded = 1") = "0"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1503016316642689026") = "0"
test $(find images -mindepth 2 | wc -l) = "0"
tw download_tweet_content https://twitter.com/wrathofgnon/status/1503016316642689026
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1503016316642689026 and is_downloaded = 0") = "0"
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1503016316642689026 and is_downloaded = 1") = "4"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1503016316642689026") = "1"
test $(find images -mindepth 2 | wc -l) = "4"

# Try to double-download it
tw fetch_tweet_only https://twitter.com/wrathofgnon/status/1503016316642689026
test $(sqlite3 twitter.db "select count(*) from tweets") = "1"
test $(sqlite3 twitter.db "select count(*) from images") = "4"


# Fetch a tweet with a video
tw fetch_user SpaceX
test $(sqlite3 twitter.db "select handle from users" | wc -l) = "2"
tw fetch_tweet_only https://twitter.com/SpaceX/status/1581025285524242432
test $(sqlite3 twitter.db "select count(*) from tweets") = "2"
test $(sqlite3 twitter.db "select count(*) from videos") = "1"

# Download the video
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1581025285524242432 and is_downloaded = 0") = "1"
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1581025285524242432 and is_downloaded = 1") = "0"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1581025285524242432") = "0"
test $(find videos -mindepth 2 | wc -l) = "0"
test $(find video_thumbnails -mindepth 2| wc -l) = "0"
tw download_tweet_content 1581025285524242432
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1581025285524242432 and is_downloaded = 0") = "0"
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1581025285524242432 and is_downloaded = 1") = "1"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1581025285524242432") = "1"
test $(find videos -mindepth 2 | wc -l) = "1"
test $(find video_thumbnails -mindepth 2 | wc -l) = "1"

# Try to double-download it
tw fetch_tweet_only https://twitter.com/SpaceX/status/1581025285524242432
test $(sqlite3 twitter.db "select count(*) from tweets") = "2"
test $(sqlite3 twitter.db "select count(*) from videos") = "1"


# Fetch a tweet with a GIF
tw fetch_user Cernovich
initial_videos_count=$(find videos -mindepth 2 | wc -l)  # Don't count prefix dirs
initial_videos_db_count=$(sqlite3 twitter.db "select count(*) from videos")
tw fetch_tweet_only https://twitter.com/Cernovich/status/1444429517020274693

test $(sqlite3 twitter.db "select count(*) from videos") = "$((initial_videos_db_count + 1))"
test $(sqlite3 twitter.db "select is_gif from videos where tweet_id = 1444429517020274693") = "1"

# Download the GIF
test $(find videos -mindepth 2 | wc -l) = "$((initial_videos_count))"   # Shouldn't have changed yet
tw download_tweet_content https://twitter.com/Cernovich/status/1444429517020274693
test $(find videos -mindepth 2 | wc -l) = "$((initial_videos_count + 1))"


# Fetch a tweet with 2 gifs
initial_videos_count=$(find videos -mindepth 2 | wc -l)
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1582197943511023616 and is_gif = 1") = "0"
tw fetch_tweet 1582197943511023616
test $(find videos -mindepth 2 | wc -l) = "$((initial_videos_count + 2))"
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1582197943511023616 and is_gif = 1") = "2"

# Fetch a tweet with 2 videos
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1591025378143129601 and is_gif = 0") = "0"
tw fetch_user alifarhat79
tw fetch_tweet_only https://twitter.com/alifarhat79/status/1591025378143129601
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1591025378143129601 and is_gif = 0") = "2"

initial_videos_count=$(find videos -mindepth 2 | wc -l)
tw download_tweet_content https://twitter.com/alifarhat79/status/1591025378143129601
test $(find videos -mindepth 2 | wc -l) = "$((initial_videos_count + 2))"

# Fetch a tweet with a video and an image
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1579292281898766336") = "0"
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1579292281898766336") = "0"
tw fetch_user mexicanwilddog
tw fetch_tweet_only https://twitter.com/mexicanwilddog/status/1579292281898766336
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1579292281898766336") = "1"
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1579292281898766336") = "1"

initial_videos_count=$(find videos -mindepth 2 | wc -l)
initial_images_count=$(find images -mindepth 2 | wc -l)
tw download_tweet_content https://twitter.com/mexicanwilddog/status/1579292281898766336
test $(find videos -mindepth 2 | wc -l) = "$((initial_videos_count + 1))"
test $(find images -mindepth 2 | wc -l) = "$((initial_images_count + 1))"


# Fetch and attempt to download a DMCAed tweet
# tw fetch_user TyCardon # TODO: This guy went private
# tw fetch_tweet_only https://twitter.com/TyCardon/status/1480640777281839106
# tw download_tweet_content 1480640777281839106
# test $(sqlite3 twitter.db "select is_blocked_by_dmca, is_downloaded from videos where tweet_id = 1480640777281839106") = "1|0"

# Fetch a tweet with a poll
tw fetch_tweet 1465534109573390348
test $(sqlite3 twitter.db "select count(*) from polls where tweet_id = 1465534109573390348") = "1"
test "$(sqlite3 twitter.db "select choice1, choice2, choice3, choice4 from polls where tweet_id = 1465534109573390348")" = "Tribal armband|Marijuana leaf|Butterfly|Maple leaf"
test "$(sqlite3 twitter.db "select choice1_votes, choice2_votes, choice3_votes, choice4_votes from polls where tweet_id = 1465534109573390348")" = "1593|624|778|1138"


# Fetch a tweet with a Twitter Space
tw fetch_tweet https://twitter.com/lndian_Bronson/status/1569875562784608256
test $(sqlite3 twitter.db "select count(*) from spaces") = "1"
test $(sqlite3 twitter.db "select space_id from tweets where id = 1569875562784608256") = "1dRJZMzeDpNGB"

# Download a full thread
tw fetch_tweet https://twitter.com/RememberAfghan1/status/1429585423702052867
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429585423702052867") = "RememberAfghan1"
test $(sqlite3 twitter.db "select is_conversation_scraped, abs(last_scraped_at - strftime('%s','now') || substr(strftime('%f','now'),4)) < 30000 from tweets where id = 1429585423702052867") = "1|1"
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429584239570391042") = "michaelmalice"
test $(sqlite3 twitter.db "select is_conversation_scraped from tweets where id = 1429584239570391042") = "0"
test "$(sqlite3 twitter.db "select handle, is_banned from tweets join users on tweets.user_id = users.id where tweets.id=1429583672827465730")" = "kanesays23|1"  # This guy got banned
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429616911315345414") = "NovaValentis"
test $(sqlite3 twitter.db "select reply_mentions from tweets where id = 1429585423702052867") = "michaelmalice"
test $(sqlite3 twitter.db "select reply_mentions from tweets where id = 1429616911315345414") = "RememberAfghan1,michaelmalice"


# Test that profile images (tiny vs regular) are chosen properly
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle = 'Cernovich'") = "1"
test $(find profile_images/Cernovich* | grep normal | wc -l) = "0"  # Since "Cernovich" was fetched directly, should have full-sized profile image and banner
test $(find profile_images/Cernovich* | grep banner | wc -l) = "1"

test $(sqlite3 twitter.db "select is_content_downloaded from users where handle = 'RememberAfghan1'") = "0"
test $(find profile_images/RememberAfghan1* | grep normal | wc -l) = "1"  # "RememberAfghan1" was fetched via a tweet thread and isn't followed, so should have tiny profile image and no banner
test $(find profile_images/RememberAfghan1* | grep banner | wc -l) = "0"


# Test that the `--profile` flag works
cd ..
tw --profile data fetch_user elonmusk
test $(sqlite3 data/twitter.db "select count(*) from users where handle = 'elonmusk'") = "1"
cd data

# Test that fetching tweets with ID only (not full URL) works
test $(sqlite3 twitter.db "select count(*) from tweets where id = 1433713164546293767") = "0"  # Check it's not already there
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle='elonmusk'") = "1"  # Should be downloaded from the previous test!
tw fetch_tweet 1433713164546293767
test $(sqlite3 twitter.db "select count(*) from tweets where id = 1433713164546293767") = "1"  # Should be there now
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle='elonmusk'") = "1"  # Should not un-set content-downloaded!


# Get a user's feed
tw fetch_user covfefeanon
covfefe_id=$(sqlite3 twitter.db "select id from users where handle like 'covfefeanon'")
test $(sqlite3 twitter.db "select count(*) from retweets") = "0"
tweet_count_1=$(sqlite3 twitter.db "select count(*) from tweets")
tw get_user_tweets covfefeanon

# Check that there are some retweets
rts_count=$(sqlite3 twitter.db "select count(*) from retweets")
test $rts_count -gt "0"

# Check that new retweets plus new tweets > 50
tweet_count_2=$(sqlite3 twitter.db "select count(*) from tweets")
test $(sqlite3 twitter.db "select count(*) from retweets where retweeted_by != $covfefe_id") = "0"
test $(($rts_count + $tweet_count_2 - $tweet_count_1)) -gt "50"


# Fetch a privated user
tw fetch_user LandsharkRides
test $(sqlite3 twitter.db "select is_private from users where handle = 'LandsharkRides'") = "1"


# Test tweets with URLs
urls_count=$(sqlite3 twitter.db "select count(*) from urls")
test "$(sqlite3 twitter.db "select * from tweets where id = 1760459421291856312")" = ""  # Check it's not already there
tw fetch_tweet_only https://twitter.com/zerohedge/status/1760459421291856312
urls_count_after=$(sqlite3 twitter.db "select count(*) from urls")
test $urls_count_after = $(($urls_count + 1))
test "$(sqlite3 twitter.db "select title from urls where tweet_id = 1760459421291856312")" = "How Do Democrats & Republicans Feel About Certain US Industries"
test $(sqlite3 twitter.db "select count(*) from urls where tweet_id = 1760459421291856312") = "1"
thumbnail_name=$(sqlite3 twitter.db "select thumbnail_remote_url from urls where tweet_id = 1760459421291856312" | grep -Po "(?<=/)[\w-]+(?=\?)")
test -n "$thumbnail_name"  # Not testing for what the thumbnail url is because it keeps changing

# Try to double-fetch it; shouldn't duplicate the URL
tw fetch_tweet_only https://twitter.com/zerohedge/status/1760459421291856312
urls_count_after_2x=$(sqlite3 twitter.db "select count(*) from urls")
test $urls_count_after_2x = $urls_count_after

# Download the link's preview image
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1760459421291856312") = "0"
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1760459421291856312") = "0"
initial_link_preview_images_count=$(find link_preview_images -mindepth 2 | wc -l)
tw download_tweet_content 1760459421291856312
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1760459421291856312") = "1"
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1760459421291856312") = "1"
test $(find link_preview_images -mindepth 2 | wc -l) = "$((initial_link_preview_images_count + 1))"
find link_preview_images | grep ${thumbnail_name}\\w*.jpg


# Test a tweet with a URL but no thumbnail
tw fetch_user Xirong7
tw fetch_tweet_only https://twitter.com/Xirong7/status/1413665734866186243
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1413665734866186243") = "0"
test $(sqlite3 twitter.db "select has_thumbnail from urls where tweet_id = 1413665734866186243") = "0"
initial_link_preview_images_count=$(find link_preview_images | wc -l)  # Check that it doesn't change, since there's no thumbnail
tw download_tweet_content 1413665734866186243
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1413665734866186243") = "1"
test $(find link_preview_images | wc -l) = $initial_link_preview_images_count  # Should be the same


# Test a tweet thread with tombstones
tw fetch_tweet https://twitter.com/CovfefeAnon/status/1454526270809726977
test $(sqlite3 twitter.db "select is_stub from tweets where id = 1454515503242829830") = 1
test $(sqlite3 twitter.db "select is_stub from tweets where id = 1454521424144654344") = 1
test $(sqlite3 twitter.db "select is_stub from tweets where id = 1454522147750260742") = 1
test $(sqlite3 twitter.db "select is_stub from tweets where id = 1454526270809726977") = 0
# Check that it downloaded the fetchable user's profile image
test $(find profile_images/itsbackwereover_profile* | wc -l) -ne 0


# Test an expanding ("Show more") tweet
tw fetch_tweet https://twitter.com/PaulSkallas/status/1649600354747572225
test $(sqlite3 twitter.db "select is_expandable from tweets where id = 1649600354747572225") = 1
test $(sqlite3 twitter.db "select length(text) from tweets where id = 1649600354747572225") -gt 280
test "$(sqlite3 twitter.db "select text from tweets where id = 1649600354747572225" | tail -n 1)" = "A fitting ending to a time not worth saving"


# Test updating a tombstone (e.g., the QT-ing user is blocked but acct is not priv)
tw fetch_tweet https://twitter.com/michaelmalice/status/1479540552081326085
test "$(sqlite3 twitter.db "select tombstone_type, text from tweets where id = 1479540319410696192")" = "3|"

tw fetch_tweet_only 1479540319410696192  # Should remove the tombstone type and update the text
test "$(sqlite3 twitter.db "select tombstone_type, text from tweets where id = 1479540319410696192")" = "|Eyyy! Look! Another one on my block list! Well done @michaelmalice, you silck person."


# Test no-clobbering of num_likes/num_retweets etc when a tweet gets deleted/tombstoned
# TODO: this tweet got deleted
# tw fetch_tweet 1489428890783461377  # Quoted tweet
# test "$(sqlite3 twitter.db "select tombstone_type from tweets where id = 1489428890783461377")" = ""  # Should not be tombstoned
# test "$(sqlite3 twitter.db "select num_likes from tweets where id = 1489428890783461377")" -gt "50"  # Should have some likes
# initial_vals=$(sqlite3 twitter.db "select num_likes, num_retweets, num_replies, num_quote_tweets from tweets where id = 1489428890783461377")
# tw fetch_tweet 1489432246452985857  # Quoting tweet
# test "$(sqlite3 twitter.db "select tombstone_type from tweets where id = 1489428890783461377")" -gt "0"  # Should be hidden
# test "$(sqlite3 twitter.db "select num_likes, num_retweets, num_replies, num_quote_tweets from tweets where id = 1489428890783461377")" = "$initial_vals"


# Test a tweet thread with a deleted account; should generate a user with a fake ID
tw fetch_tweet https://twitter.com/CovfefeAnon/status/1365278017233313795
test $(sqlite3 twitter.db "select is_id_fake from users where handle = '_selfoptimizer'") = 1
test $(sqlite3 twitter.db "select count(*) from tweets where user_id = (select id from users where handle = '_selfoptimizer')") = 1


# Test fetching a banned user
tw fetch_user nancytracker
test "$(sqlite3 twitter.db "select is_content_downloaded, is_banned from users where handle='nancytracker'")" = "1|1"


# Fetch a user with "600x200" banner image
tw fetch_user AlexKoppelman  # This is probably kind of a flimsy test
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle='AlexKoppelman'") = "1"


# Test following / unfollowing a user
test "$(sqlite3 twitter.db "select count(*) from users where is_followed = 1")" = "0"
tw follow michaelmalice
test "$(sqlite3 twitter.db "select handle from users where is_followed = 1")" = "michaelmalice"

tw follow cernovich
test $(tw list_followed | wc -l) = 2
test "$(tw list_followed | grep -iq cernovich && echo YES)" = "YES"
test "$(tw list_followed | grep -iq michaelmalice && echo YES)" = "YES"
test "$(tw list_followed | grep -iq blahblahgibberish && echo YES)" = ""

tw unfollow michaelmalice
test "$(sqlite3 twitter.db "select count(*) from users where is_followed = 1")" = "1"
tw unfollow cernovich
test "$(sqlite3 twitter.db "select count(*) from users where is_followed = 1")" = "0"


# When not logged in, age-restricted tweet should fail to fetch
tw fetch_user PandasAndVidya
tw fetch_tweet_only https://twitter.com/PandasAndVidya/status/1562714727968428032 || true  # This one is expected to fail
test "$(sqlite3 twitter.db "select count(*) from tweets where id = 156271472796842803")" == "0"

# Fetch an age-restricted tweet while logged in
tw fetch_tweet_only https://twitter.com/PandasAndVidya/status/1562714727968428032
test "$(sqlite3 twitter.db "select count(*) from tweets where id = 156271472796842803")" == "0"

# Test that you can pass a session with the `.session` file extension too
tw --session Offline_Twatter.session list_followed > /dev/null  # Dummy operation


# Test search
tw  search "from:michaelmalice constitution"
# Update 2024-05-30: the default search page doesn't paginate anymore
test $(sqlite3 twitter.db "select count(*) from tweets where user_id = 44067298 and text like '%constitution%'") -gt "20"  # Not sure exactly how many


# Test fetching user Likes
tw fetch_user Offline_Twatter
tw get_user_likes Offline_Twatter
test $(sqlite3 twitter.db "select count(*) from likes") -ge "2"
test $(sqlite3 twitter.db "select count(*) from likes where tweet_id = 1671902735250124802") = "1"


# Test liking and unliking
tw fetch_tweet_only 1589023388676554753
test $(sqlite3 twitter.db "select count(*) from likes where tweet_id = 1589023388676554753 and user_id = (select id from users where handle like 'offline_twatter')") = "0"
tw like_tweet https://twitter.com/elonmusk/status/1589023388676554753
test $(sqlite3 twitter.db "select count(*) from likes where tweet_id = 1589023388676554753 and user_id = (select id from users where handle like 'offline_twatter')") = "1"
tw unlike_tweet https://twitter.com/elonmusk/status/1589023388676554753
# TODO: implement deleting a Like
# test $(sqlite3 twitter.db "select count(*) from likes where tweet_id = 1589023388676554753 and user_id = (select id from users where handle like 'offline_twatter')") = "0"

# Test fetching bookmarks
tw get_bookmarks
test $(sqlite3 twitter.db "select count(*) from bookmarks") -ge "2"
test $(sqlite3 twitter.db "select count(*) from bookmarks where tweet_id = 1762239926437843421") = "1"

# Test fetch inbox
test $(sqlite3 twitter.db "select count(*) from chat_rooms") = "0"
test $(sqlite3 twitter.db "select count(*) from chat_messages") = "0"
tw fetch_inbox
test $(sqlite3 twitter.db "select count(*) from chat_rooms") -ge "1"
test $(sqlite3 twitter.db "select count(*) from chat_messages where chat_room_id = '1458284524761075714-1488963321701171204'") -ge "5"


# Test fetch a DM conversation
tw fetch_dm "1458284524761075714-1488963321701171204"


# Test followers and followees
test $(sqlite3 twitter.db "select count(*) from follows") = "0"
tw get_followees Offline_Twatter
test $(sqlite3 twitter.db "select count(*) from follows where follower_id = 1488963321701171204") = "4"
test $(sqlite3 twitter.db "select count(*) from follows where followee_id = 1488963321701171204") = "0"
tw get_followers Offline_Twatter
test $(sqlite3 twitter.db "select count(*) from follows where follower_id = 1488963321701171204 and followee_id = 759251") = "1"

# TODO: Maybe this file should be broken up into multiple test scripts

echo -e "\033[32mAll tests passed.  Finished successfully.\033[0m"
