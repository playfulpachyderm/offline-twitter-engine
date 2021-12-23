# Release Process

- update CHANGELOG.txt
- if `persistence/schema.sql` has changed since last release:
	- bump ENGINE_DATABASE_VERSION in `persistence/versions.go`
	- add an entry to MIGRATIONS in `persistence/versions.go`
