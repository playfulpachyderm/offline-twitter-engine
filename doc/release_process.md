# Release Process

- update CHANGELOG.txt
- if `persistence/schema.sql` has changed since last release:
	- add an entry to MIGRATIONS in `persistence/versions.go`
	- update `sample_data/seed_data.sql`, including database_version
- checkout a branch `vX.Y.Z` with the version number X.Y.Z and push it
