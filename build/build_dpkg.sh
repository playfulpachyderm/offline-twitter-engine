#!/bin/bash

set -e

if [[ -z "$1" ]]
then
	# Error message and exit
	>&2 echo "No version number provided!  Exiting."
	exit 1
fi

# Compile the program
(cd ../cmd && ./compile.sh)

# Prepare the output folder
if [[ -e dpkg_tmp ]]
then
	rm -rf dpkg_tmp
fi
mkdir dpkg_tmp

# Construct the dpkg directory structure
mkdir -p dpkg_tmp/usr/local/bin
cp ../cmd/tw dpkg_tmp/usr/local/bin/twitter

# Create the `DEBIAN/control` file
mkdir dpkg_tmp/DEBIAN
echo "Package: offline-twitter-engine
Version: $1
Architecture: all
Maintainer: me@playfulpachyderm.com
Installed-Size: 7200
Depends:
Section: web
Priority: optional
Homepage: http://offline-twitter.com
Description: This utility is the scraper engine that drives \`offline-twitter\`.
 Download and browse content from twitter.
 Save a local copy of everything you browse to a SQLite database.
" > dpkg_tmp/DEBIAN/control


dpkg-deb --build `pwd`/dpkg_tmp .