fumig
=====

This is a simple gopher browser. It's a quick and nasty hack, so the
code is a mess for now. It runs in a terminal but will open links
(v) with `xdg-open`, `start` or `open` depending on your OS.

Controls
--------

  Key                Action
  ------------------ ----------------------------------------------
  o                  Open URI
  q                  Quit
  h,left,backspace   Go back in history
  l,right,return     Enter menu or download file under cursor
  j,down             Move cursor down
  k,up               Move cursor up
  J,n                Move to next menu item
  K,p                Move to previous menu item
  d,space,pgdn       Move cursor down half a page
  u,pgup             Move cursor up half a page
  g,home             Move cursor to start of document
  G,end              Move cursor to end of document
  v                  View file under cursor
  r,F5               Reload current page
  ^L                 Redraw UI

