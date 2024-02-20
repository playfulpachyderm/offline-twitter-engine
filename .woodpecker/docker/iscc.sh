#!/bin/sh

# Stolen from:
# https://github.com/amake/innosetup-docker under `opt/bin/iscc`

set -eu

escone() {
    printf %s\\n "$1" | sed "s/'/'\\\\''/g;1s/^/'/;\$s/\$/' \\\\/"
}

winpaths() {
    for arg; do
        if [ -e "$arg" ]; then
            escone "$(winepath -w "$arg")"
        else
            escone "$arg"
        fi
    done
    echo " "
}

# PROGFILES_PATH="$(winepath -u "$(wine cmd /c "echo %PROGRAMFILES%" | tr -d '\r')")"
PROGFILES_PATH="$(winepath -u "C:\Program Files")"

# Set args (`$@`) to the map of `winpaths` over the existing args
eval set -- "$(winpaths "$@")"

exec wine "${PROGFILES_PATH}/Inno Setup 6/ISCC.exe" "$@"
