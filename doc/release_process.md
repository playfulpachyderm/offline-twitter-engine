# Release Process

- update CHANGELOG.txt
- if `persistence/schema.sql` has changed since last release:
	- add an entry to MIGRATIONS in `persistence/versions.go`
	- update `sample_data/seed_data.sql`, including database_version
- checkout a branch `release-x.y.z` with the version number x.y.z and push it
