# Release Process

- update CHANGELOG.txt
- if `persistence/schema.sql` has changed since last release:
	- add an entry to MIGRATIONS in `persistence/versions.go`
- checkout a branch `release-x.y.z` with the version number x.y.z and push it
