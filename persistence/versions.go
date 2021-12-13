package persistence

import (
	"fmt"
	"database/sql"

	"offline_twitter/terminal_utils"
)


const ENGINE_DATABASE_VERSION = 1


type VersionMismatchError struct {
	EngineVersion int
	DatabaseVersion int
}
func (e VersionMismatchError) Error() string {
	return fmt.Sprintf(
`This profile was created with database schema version %d, which is newer than this application's database schema version, %d.
Please upgrade this application to a newer version to use this profile.  Or downgrade the profile's schema version, somehow.`,
			e.DatabaseVersion, e.EngineVersion,
	)
}


/**
 * The Nth entry is the migration that moves you from version N to version N+1
 */
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
}

/**
 * This should only get called on a newly created Profile.
 * Subsequent updates should change the number, not insert a new row.
 */
func InitializeDatabaseVersion(db *sql.DB) {
	_, err := db.Exec("insert into database_version (version_number) values (?)", ENGINE_DATABASE_VERSION)
	if err != nil {
		panic(err)
	}
}

func (p Profile) GetDatabaseVersion() (int, error) {
	row := p.DB.QueryRow("select version_number from database_version")

	var version int

	err := row.Scan(&version)
	if err != nil {
		return 0, err
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
		fmt.Printf("Database version is out of date.  Upgrading database from version v%d to version v%d!\n", version, ENGINE_DATABASE_VERSION)
		fmt.Printf(terminal_utils.COLOR_RESET)
		return p.UpgradeFromXToY(version, ENGINE_DATABASE_VERSION)
	}

	return nil
}

/**
 * Run all the migrations from version X to version Y, and update the `database_version` table's `version_number`
 */
func (p Profile) UpgradeFromXToY(x int, y int) error {
	for i := x; i < y; i++ {
		fmt.Printf(terminal_utils.COLOR_CYAN)
		fmt.Println(MIGRATIONS[i])
		fmt.Printf(terminal_utils.COLOR_RESET)

		_, err := p.DB.Exec(MIGRATIONS[i])
		if err != nil {
			return err
		}
		_, err = p.DB.Exec("update database_version set version_number = ?", i+1)
		if err != nil {
			return err
		}
		fmt.Printf(terminal_utils.COLOR_YELLOW)
		fmt.Printf("Now at database schema version %d.\n", i + 1)
		fmt.Printf(terminal_utils.COLOR_RESET)
	}
	fmt.Printf(terminal_utils.COLOR_GREEN)
	fmt.Printf("================================================\n")
	fmt.Printf("Database version has been upgraded to version %d.\n", y)
	fmt.Printf(terminal_utils.COLOR_RESET)
	return nil
}
