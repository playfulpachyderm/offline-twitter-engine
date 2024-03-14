# Inspired by: https://github.com/amake/innosetup-docker/

from ubuntu:jammy
shell ["/bin/bash", "-c"]
run dpkg --add-architecture i386
run apt update
run apt install -y curl ssh wine wine32 xvfb winbind

run curl -SL "https://files.jrsoftware.org/is/6/innosetup-6.2.2.exe" -o is.exe
env DISPLAY ":99"
env WINEDEBUG "-all,err+all"
run xvfb-run wine is.exe /SP- /VERYSILENT /ALLUSERS /SUPPRESSMSGBOXES /DOWNLOADISCRYPT=1
copy iscc.sh /usr/bin/iscc.sh

# For SSH upload
copy known_hosts /root/.ssh/known_hosts
