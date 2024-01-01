package persistence

import (
	"fmt"

	sql "github.com/jmoiron/sqlx"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/terminal_utils"
)

type VersionMismatchError struct {
	EngineVersion   int
	DatabaseVersion int
}

func (e VersionMismatchError) Error() string {
	return fmt.Sprintf(
		`This profile was created with database schema version %d, which is newer than this application's database schema version, %d.
Please upgrade this application to a newer version to use this profile.  Or downgrade the profile's schema version, somehow.`,
		e.DatabaseVersion, e.EngineVersion,
	)
}

// The Nth entry is the migration that moves you from version N to version N+1.
var MIGRATIONS = []string{
	`create table polls (rowid integer primary key,
		    id integer unique not null check(typeof(id) = 'integer'),
		    tweet_id integer not null,
		    num_choices integer not null,

		    choice1 text,
		    choice1_votes integer,
		    choice2 text,
		    choice2_votes integer,
		    choice3 text,
		    choice3_votes integer,
		    choice4 text,
		    choice4_votes integer,

		    voting_duration integer not null,  -- in seconds
		    voting_ends_at integer not null,

		    last_scraped_at integer not null,

		    foreign key(tweet_id) references tweets(id)
		);`,
	`alter table tweets add column is_conversation_scraped boolean default 0;
		alter table tweets add column last_scraped_at integer not null default 0`,
	`update tombstone_types set tombstone_text = 'This Tweet is from a suspended account' where rowid = 2;
		insert into tombstone_types (rowid, short_name, tombstone_text)
		                     values (5, 'violated', 'This Tweet violated the Twitter Rules'),
		                            (6, 'no longer exists', 'This Tweet is from an account that no longer exists')`,
	`alter table videos add column thumbnail_remote_url text not null default "missing";
		alter table videos add column thumbnail_local_filename text not null default "missing"`,

	// 5
	`alter table videos add column duration integer not null default 0;
		alter table videos add column view_count integer not null default 0`,
	`alter table users add column is_banned boolean default 0`,
	`alter table urls add column short_text text not null default ""`,
	`insert into tombstone_types (rowid, short_name, tombstone_text) values (7, 'age-restricted', 'Age-restricted adult content. '
		|| 'This content might not be appropriate for people under 18 years old. To view this media, you’ll need to log in to Twitter')`,
	`alter table users add column is_followed boolean default 0`,

	// 10
	`create table fake_user_sequence(latest_fake_id integer not null);
		insert into fake_user_sequence values(0x4000000000000000);
		alter table users add column is_id_fake boolean default 0;`,
	`delete from urls where rowid in (select urls.rowid from tweets join urls on tweets.id = urls.tweet_id where urls.text like
		'https://twitter.com/%/status/' || tweets.quoted_tweet_id || "%")`,
	`create table spaces(rowid integer primary key,
		    id text unique not null,
		    short_url text not null
		);
		alter table tweets add column space_id text references spaces(id)`,
	`alter table videos add column is_blocked_by_dmca boolean not null default 0`,
	`create index if not exists index_tweets_in_reply_to_id on tweets (in_reply_to_id);
		create index if not exists index_urls_tweet_id on urls (tweet_id);
		create index if not exists index_polls_tweet_id on polls (tweet_id);
		create index if not exists index_images_tweet_id on images (tweet_id);
		create index if not exists index_videos_tweet_id on videos (tweet_id);`,

	// 15
	`alter table spaces add column created_by_id integer references users(id);
		alter table spaces add column state text not null default "";
		alter table spaces add column title text not null default "";
		alter table spaces add column created_at integer;
		alter table spaces add column started_at integer;
		alter table spaces add column ended_at integer;
		alter table spaces add column updated_at integer;
		alter table spaces add column is_available_for_replay boolean not null default 0;
		alter table spaces add column replay_watch_count integer;
		alter table spaces add column live_listeners_count integer;
		alter table spaces add column is_details_fetched boolean not null default 0;
		create table space_participants(rowid integer primary key,
		    user_id integer not null,
		    space_id not null,
		    foreign key(space_id) references spaces(id)
		);`,
	`create index if not exists index_tweets_user_id on tweets (user_id);`,
	`alter table tweets add column is_expandable bool not null default 0;`,
	`create table space_participants_uniq(rowid integer primary key,
			user_id integer not null,
			space_id not null,

			unique(user_id, space_id)
			foreign key(space_id) references spaces(id)
			-- No foreign key for users, since they may not be downloaded yet and I don't want to
			-- download every user who joins a space
		);

		insert or replace into space_participants_uniq(rowid, user_id, space_id) select rowid, user_id, space_id from space_participants;

		drop table space_participants;
		alter table space_participants_uniq rename to space_participants;
		vacuum;`,
	`create table likes(rowid integer primary key,
			sort_order integer unique not null,
			user_id integer not null,
			tweet_id integer not null,
			unique(user_id, tweet_id)
			foreign key(user_id) references users(id)
			foreign key(tweet_id) references tweets(id)
		);`,

	// 20
	`create index if not exists index_tweets_posted_at on tweets (posted_at);
		create index if not exists index_retweets_retweeted_at on retweets (retweeted_at)`,
	`update spaces set ended_at = ended_at/1000 where ended_at > strftime("%s")*500;
		update spaces set updated_at = updated_at/1000 where updated_at > strftime("%s")*500;
		update spaces set started_at = started_at/1000 where started_at > strftime("%s")*500;
		update spaces set created_at = created_at/1000 where created_at > strftime("%s")*500;`,
	`alter table users add column is_deleted boolean default 0`,
	`begin transaction;
		alter table likes rename to likes_old;

		create table likes(rowid integer primary key,
		    sort_order integer not null,
		    user_id integer not null,
		    tweet_id integer not null,
		    unique(user_id, tweet_id)
		    foreign key(user_id) references users(id)
		    foreign key(tweet_id) references tweets(id)
		);

		create index if not exists index_likes_user_id on likes (user_id);
		create index if not exists index_likes_tweet_id on likes (tweet_id);
		insert into likes select * from likes_old;
		drop table likes_old;
		commit;
		vacuum;`,
	`insert into tombstone_types(rowid, short_name, tombstone_text)
	                     values (8, 'newer-version-available', 'There’s a new version of this Tweet')`,

	// 25
	`create table chat_rooms (rowid integer primary key,
		    id text unique not null,
		    type text not null,
		    last_messaged_at integer not null,
		    is_nsfw boolean not null,

		    -- Group DM info
		    created_at integer not null,
		    created_by_user_id integer not null,
		    name text not null default '',
		    avatar_image_remote_url text not null default '',
		    avatar_image_local_path text not null default ''
		);

		create table chat_room_participants(rowid integer primary key,
		    chat_room_id text not null,
		    user_id integer not null,
		    last_read_event_id integer not null,
		    is_chat_settings_valid boolean not null default 0,
		    is_notifications_disabled boolean not null,
		    is_mention_notifications_disabled boolean not null,
		    is_read_only boolean not null,
		    is_trusted boolean not null,
		    is_muted boolean not null,
		    status text not null,
		    unique(chat_room_id, user_id)
		);

		create table chat_messages (rowid integer primary key,
		    id integer unique not null check(typeof(id) = 'integer'),
		    chat_room_id text not null,
		    sender_id integer not null,
		    sent_at integer not null,
		    request_id text not null,
		    in_reply_to_id integer,
		    text text not null,
		    embedded_tweet_id integer not null default 0,
		    foreign key(chat_room_id) references chat_rooms(id)
		    foreign key(sender_id) references users(id)
		);

		create table chat_message_reactions (rowid integer primary key,
		    id integer unique not null check(typeof(id) = 'integer'),
		    message_id integer not null,
		    sender_id integer not null,
		    sent_at integer not null,
		    emoji text not null,
		    foreign key(message_id) references chat_messages(id)
		    foreign key(sender_id) references users(id)
		);`,
	`create table follows(rowid integer primary key,
		    follower_id integer not null,
		    followee_id integer not null,
		    unique(follower_id, followee_id),
		    foreign key(follower_id) references users(id)
		    foreign key(followee_id) references users(id)
		);
		create index if not exists index_follows_followee_id on follows (followee_id);
		create index if not exists index_follows_follower_id on follows (follower_id);`,
}
var ENGINE_DATABASE_VERSION = len(MIGRATIONS)

// This should only get called on a newly created Profile.
// Subsequent updates should change the number, not insert a new row.
func InitializeDatabaseVersion(db *sql.DB) {
	db.MustExec("insert into database_version (version_number) values (?)", ENGINE_DATABASE_VERSION)
}

func (p Profile) GetDatabaseVersion() (int, error) {
	row := p.DB.QueryRow("select version_number from database_version")

	var version int

	err := row.Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("Error checking database version:\n  %w", err)
	}
	return version, nil
}

func (p Profile) check_and_update_version() error {
	version, err := p.GetDatabaseVersion()
	if err != nil {
		return err
	}

	if version > ENGINE_DATABASE_VERSION {
		return VersionMismatchError{ENGINE_DATABASE_VERSION, version}
	}

	if ENGINE_DATABASE_VERSION > version {
		fmt.Printf(terminal_utils.COLOR_YELLOW)
		fmt.Printf("================================================\n")
		fmt.Printf("Database version is out of date.  Upgrading database from version %d to version %d!\n", version,
			ENGINE_DATABASE_VERSION)
		fmt.Printf(terminal_utils.COLOR_RESET)
		return p.UpgradeFromXToY(version, ENGINE_DATABASE_VERSION)
	}

	return nil
}

// Run all the migrations from version X to version Y, and update the `database_version` table's `version_number`
func (p Profile) UpgradeFromXToY(x int, y int) error {
	for i := x; i < y; i++ {
		fmt.Printf(terminal_utils.COLOR_CYAN)
		fmt.Println(MIGRATIONS[i])
		fmt.Printf(terminal_utils.COLOR_RESET)

		p.DB.MustExec(MIGRATIONS[i])
		p.DB.MustExec("update database_version set version_number = ?", i+1)

		fmt.Printf(terminal_utils.COLOR_YELLOW)
		fmt.Printf("Now at database schema version %d.\n", i+1)
		fmt.Printf(terminal_utils.COLOR_RESET)
	}
	fmt.Printf(terminal_utils.COLOR_GREEN)
	fmt.Printf("================================================\n")
	fmt.Printf("Database version has been upgraded to version %d.\n", y)
	fmt.Printf(terminal_utils.COLOR_RESET)
	return nil
}
