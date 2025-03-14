import sys
import sqlite3
import os

tweet_id = sys.argv[1]

db = sqlite3.connect("twitter.db")
c = db.cursor()

# Get video ID and filepath
c.execute("select id, local_filename from videos where tweet_id = ?", (tweet_id, ))
for id, local_filename in c.fetchall():
	# Mark the video as not-downloaded
	c.execute("update videos set is_downloaded=0 where id = ?", (id, ))
	# Delete the video on disk
	os.remove("videos/" + local_filename)

# Mark the tweet as not downloaded
c.execute("update tweets set is_content_downloaded=0 where id = ?", (tweet_id, ))

db.commit()
