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
test $(find profile_images | wc -l) = "1"                      # should be empty to begin
tw fetch_user Denlesks
test "$(sqlite3 twitter.db "select handle from users")" = "Denlesks"
test $(sqlite3 twitter.db "select count(*) from users") = "1"
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle = 'Denlesks'") = "1"
test $(find profile_images | wc -l) = "3"                      # should have gotten 2 images
test -f profile_images/Denlesks_profile_22YJvhC7.jpg
test -f profile_images/Denlesks_banner_1585776052.jpg
tw fetch_user Denlesks                                         # try to double-download it
test $(sqlite3 twitter.db "select count(*) from users") = "1"  # shouldn't have added a new row


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
tw download_tweet_content https://twitter.com/Denlesks/status/1261483383483293700
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


# Fetch a tweet with a GIF
tw fetch_user Cernovich
initial_videos_count=$(find videos | wc -l)
initial_videos_db_count=$(sqlite3 twitter.db "select count(*) from videos")
tw fetch_tweet_only https://twitter.com/Cernovich/status/1444429517020274693

test $(sqlite3 twitter.db "select count(*) from videos") = "$((initial_videos_db_count + 1))"
test $(sqlite3 twitter.db "select is_gif from videos where tweet_id = 1444429517020274693") = "1"

# Download the GIF
test $(find videos | wc -l) = "$((initial_videos_count))"   # Shouldn't have changed yet
tw download_tweet_content https://twitter.com/Cernovich/status/1444429517020274693
test $(find videos | wc -l) = "$((initial_videos_count + 1))"



# Download a full thread
tw fetch_tweet https://twitter.com/RememberAfghan1/status/1429585423702052867
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429585423702052867") = "RememberAfghan1"
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429584239570391042") = "michaelmalice"
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429583672827465730") = "kanesays23"
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429616911315345414") = "NovaValentis"
test $(sqlite3 twitter.db "select reply_mentions from tweets where id = 1429585423702052867") = "michaelmalice"
test $(sqlite3 twitter.db "select reply_mentions from tweets where id = 1429616911315345414") = "RememberAfghan1,michaelmalice"


# Test that the `--profile` flag works
cd ..
tw --profile data fetch_user elonmusk
test $(sqlite3 data/twitter.db "select count(*) from users where handle = 'elonmusk'") = "1"
cd data

# Test that fetching tweets with ID only (not full URL) works
test $(sqlite3 twitter.db "select count(*) from tweets where id = 1433713164546293767") = "0"  # Check it's not already there
tw fetch_tweet 1433713164546293767
test $(sqlite3 twitter.db "select count(*) from tweets where id = 1433713164546293767") = "1"  # Should be there now


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


# Test tweets with URLs
urls_count=$(sqlite3 twitter.db "select count(*) from urls")
tw fetch_tweet https://twitter.com/CovfefeAnon/status/1428904664645394433
urls_count_after=$(sqlite3 twitter.db "select count(*) from urls")
test $urls_count_after = $(($urls_count + 1))
test "$(sqlite3 twitter.db "select title from urls where tweet_id = 1428904664645394433")" = "Justice Department investigating Elon Musk's SpaceX following complaint of hiring discrimination"
test $(sqlite3 twitter.db "select thumbnail_remote_url from urls where tweet_id = 1428904664645394433") = "https://pbs.twimg.com/card_img/1439840148280123394/LdQZA_2E?format=jpg&name=800x320_1"

# Try to double-fetch it; shouldn't duplicate the URL
tw fetch_tweet https://twitter.com/CovfefeAnon/status/1428904664645394433
urls_count_after_2x=$(sqlite3 twitter.db "select count(*) from urls")
test $urls_count_after_2x = $urls_count_after

# Download the link's preview image
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1428904664645394433") = "0"
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1428904664645394433") = "0"
test $(find link_preview_images | wc -l) = "1"
tw download_tweet_content 1428904664645394433
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1428904664645394433") = "1"
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1428904664645394433") = "1"
test $(find link_preview_images | wc -l) = "2"
test -f link_preview_images/LdQZA_2E_800x320_1.jpg


# Test a tweet with a URL but no thumbnail
tw fetch_tweet https://twitter.com/Xirong7/status/1413665734866186243
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1413665734866186243") = "0"
test $(sqlite3 twitter.db "select has_thumbnail from urls where tweet_id = 1413665734866186243") = "0"
test $(find link_preview_images | wc -l) = "2"
tw download_tweet_content 1413665734866186243
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1413665734866186243") = "1"
test $(find link_preview_images | wc -l) = "2"



# TODO: Maybe this file should be broken up into multiple test scripts

echo -e "\033[32mAll tests passed.  Finished successfully.\033[0m"
