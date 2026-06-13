#!/usr/bin/fish

set IN "the_conservative_frame_r4.md"
set OUT "the_conservative_frame"

pandoc -o $OUT.pdf $IN
pandoc -o $OUT.html $IN
pandoc -o $OUT.epub $IN
