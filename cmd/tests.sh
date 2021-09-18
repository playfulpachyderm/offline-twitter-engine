#!/bin/bash

set -e
set -x

PS4='+(${BASH_SOURCE}:${LINENO}): '

./compile.sh

test -e data && rm -r data

PATH=`pwd`:$PATH

tw create_profile data
cd data


# Fetch a user
tw fetch_user Denlesks
test "$(sqlite3 twitter.db "select handle from users")" = "Denlesks"
test $(sqlite3 twitter.db "select count(*) from users") = "1"
tw fetch_user Denlesks
test $(sqlite3 twitter.db "select count(*) from users") = "1"


# Fetch a tweet with images
tw fetch_tweet_only https://twitter.com/Denlesks/status/1261483383483293700
test $(sqlite3 twitter.db "select count(*) from tweets") = "1"
test "$(sqlite3 twitter.db "select text from tweets")" = "These are public health officials who are making decisions about your lifestyle because they know more about health, fitness and well-being than you do"
test $(sqlite3 twitter.db "select count(*) from images") = "4"

# Download its images
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 0") = "4"
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 1") = "0"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1261483383483293700") = "0"
test $(find images | wc -l) = "1"
tw download_tweet_content 1261483383483293700
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 0") = "0"
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 1") = "4"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1261483383483293700") = "1"
test $(find images | wc -l) = "5"

# Try to double-download it
tw fetch_tweet_only https://twitter.com/Denlesks/status/1261483383483293700
test $(sqlite3 twitter.db "select count(*) from tweets") = "1"
test $(sqlite3 twitter.db "select count(*) from images") = "4"


# Fetch a tweet with a video
tw fetch_user DiamondChariots
test $(sqlite3 twitter.db "select handle from users" | wc -l) = "2"
tw fetch_tweet_only https://twitter.com/DiamondChariots/status/1418971605674467340
test $(sqlite3 twitter.db "select count(*) from tweets") = "2"
test $(sqlite3 twitter.db "select count(*) from videos") = "1"

# Download the video
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 0") = "1"
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 1") = "0"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1418971605674467340") = "0"
test $(find videos| wc -l) = "1"
tw download_tweet_content 1418971605674467340
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 0") = "0"
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 1") = "1"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1418971605674467340") = "1"
test $(find videos | wc -l) = "2"

# Try to double-download it
tw fetch_tweet_only https://twitter.com/DiamondChariots/status/1418971605674467340
test $(sqlite3 twitter.db "select count(*) from tweets") = "2"
test $(sqlite3 twitter.db "select count(*) from videos") = "1"


# Download a user's profile image and banner image
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle = 'DiamondChariots'") = "0"
tw download_user_content DiamondChariots
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle = 'DiamondChariots'") = "1"
test -f profile_images/DiamondChariots_profile_rE4OTedS.jpg
test -f profile_images/DiamondChariots_banner_1615811094.jpg


# Download a full thread
tw fetch_tweet https://twitter.com/RememberAfghan1/status/1429585423702052867
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429585423702052867") = "RememberAfghan1"
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429584239570391042") = "michaelmalice"
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429583672827465730") = "kanesays23"
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429616911315345414") = "NovaValentis"


# Test that the `--profile` flag works
cd ..
tw --profile data fetch_user elonmusk
test $(sqlite3 data/twitter.db "select count(*) from users where handle = 'elonmusk'") = "1"
cd data


# Get a user's feed
malice_id=$(sqlite3 twitter.db "select id from users where handle='michaelmalice'")
test $(sqlite3 twitter.db "select count(*) from retweets") = "0"
tweet_count_1=$(sqlite3 twitter.db "select count(*) from tweets")
tw get_user_tweets michaelmalice

# Check that there are some retweets
rts_count=$(sqlite3 twitter.db "select count(*) from retweets")
test $rts_count -gt "0"

# Check that new retweets plus new tweets > 50
tweet_count_2=$(sqlite3 twitter.db "select count(*) from tweets")
test $(sqlite3 twitter.db "select count(*) from retweets where retweeted_by != $malice_id") = "0"
test $(($rts_count + $tweet_count_2 - $tweet_count_1)) -gt "50"


# Fetch a privated user
tw fetch_user HbdNrx
test $(sqlite3 twitter.db "select is_private from users where handle = 'HbdNrx'") = "1"


echo -e "\033[32mAll tests passed.  Finished successfully.\033[0m"
