#!/bin/sh
# The .deb's postinstall and postremove script: refreshes the app list and the
# icon cache, so the app appears and disappears at once. Either tool may be
# missing (a minimal system, a container), and that's no reason to fail.
if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database -q /usr/share/applications
fi
if command -v gtk-update-icon-cache >/dev/null 2>&1; then
  gtk-update-icon-cache -q -t -f /usr/share/icons/hicolor
fi
