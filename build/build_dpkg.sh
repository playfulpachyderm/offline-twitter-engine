#!/bin/bash

set -e

if [[ -z "$1" ]]
then
	# Error message and exit
	>&2 echo "No version number provided!  Exiting."
	exit 1
fi
version=$1

(cd ../cmd && ./compile.sh --static $version)

# Prepare the output folder
if [[ -e dpkg_tmp ]]
then
	rm -rf dpkg_tmp
fi
mkdir dpkg_tmp

# Construct the dpkg directory structure
mkdir -p dpkg_tmp/usr/local/bin
cp ../cmd/tw dpkg_tmp/usr/local/bin/twitter

# Create the Desktop files
mkdir -p dpkg_tmp/usr/share/pixmaps
cp twitter.png dpkg_tmp/usr/share/pixmaps

mkdir -p dpkg_tmp/usr/share/applications
cp Twitter.desktop dpkg_tmp/usr/share/applications/Twitter.desktop



# Create the `DEBIAN/control` file
mkdir dpkg_tmp/DEBIAN
echo "Package: offline-twitter
Version: $version
Architecture: all
Maintainer: alex@playfulpachyderm.com
Installed-Size: `du -k dpkg_tmp | tail -n 1 | cut -f 1`
Depends:
Section: web
Priority: optional
Homepage: http://offline-twitter.com
Description: This utility is the scraper engine that drives \`offline-twitter\`.
 Download and browse content from twitter.  Everything you browse gets saved as
 a local copy into a SQLite database on your computer.
" > dpkg_tmp/DEBIAN/control


dpkg-deb --build `pwd`/dpkg_tmp .
