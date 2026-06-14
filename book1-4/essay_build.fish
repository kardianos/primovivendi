#!/usr/bin/fish

# Build the standalone essay in three formats.
#   the_conservative_frame.pdf  - print, via LaTeX (geometry from the md front matter)
#   the_conservative_frame.html - single-file web edition, book typography embedded
#   the_conservative_frame.epub - reflowable ebook, chapter-split, e-reader CSS
#
# Title / author / date / geometry are read from the YAML front matter of $IN.

set script_dir (dirname (status --current-filename))
cd $script_dir

set IN  "the_conservative_frame_r4.md"
set OUT "the_conservative_frame"

# ---- web (HTML) stylesheet: a centered reading column that reads like a page ----
# Justified, first-line-indented body; headings never hyphenate; styled title
# block + table of contents; respects the reader's light/dark preference.
set HTMLCSS (mktemp --suffix=.css)
printf '%s\n' \
  ':root{--measure:34em;--ink:#1a1a1a;--paper:#fdfdfb;--muted:#6a6a6a;--link:#0b3d91;}' \
  '*{box-sizing:border-box;}' \
  'html{font-size:112.5%;}' \
  'body{font-family:"Iowan Old Style","Palatino Linotype","Book Antiqua",Palatino,Georgia,Cambria,serif;line-height:1.6;color:var(--ink);background:var(--paper);max-width:var(--measure);margin:0 auto;padding:3.5rem 1.25rem 6rem;text-rendering:optimizeLegibility;-webkit-font-smoothing:antialiased;}' \
  'p{margin:0;text-align:justify;hyphens:auto;}' \
  'p + p{text-indent:1.4em;}' \
  'h1,h2,h3,h4{font-weight:700;line-height:1.25;hyphens:none;}' \
  'h2{margin:2.4em 0 .6em;font-size:1.5rem;}' \
  'h3{margin:1.8em 0 .5em;font-size:1.2rem;}' \
  'h2 + p,h3 + p,h4 + p{text-indent:0;}' \
  'header#title-block-header{text-align:center;margin:2rem 0 3rem;padding-bottom:2rem;border-bottom:1px solid var(--muted);}' \
  'h1.title{font-size:2.4rem;line-height:1.15;margin:0 0 .4em;}' \
  'p.author{font-variant:small-caps;letter-spacing:.04em;color:var(--muted);margin:.3em 0 0;}' \
  'p.date{color:var(--muted);margin:.2em 0 0;}' \
  'nav#TOC{margin:0 0 3rem;font-size:.95rem;}' \
  'nav#TOC a{text-decoration:none;color:inherit;}' \
  'nav#TOC a:hover{text-decoration:underline;}' \
  'nav#TOC ul{list-style:none;padding-left:1.2em;}' \
  'blockquote{margin:1.2em 1.6em;font-style:italic;border-left:2px solid var(--muted);padding-left:1em;}' \
  'ul,ol{margin:1em 0;padding-left:1.6em;}' \
  'li{margin:.3em 0;}' \
  'hr{border:0;margin:2.5em 0;text-align:center;}' \
  'hr::after{content:"\2766";color:var(--muted);}' \
  'a{color:var(--link);}' \
  'code{font-family:ui-monospace,Menlo,Consolas,monospace;font-size:.9em;}' \
  '@media (prefers-color-scheme:dark){:root{--ink:#e8e6e1;--paper:#16161a;--muted:#9a978f;--link:#9db8ff;}}' \
  '@media print{body{max-width:none;background:#fff;color:#000;}}' \
  > $HTMLCSS

# ---- ebook (EPUB) stylesheet: conservative, e-reader friendly ----
# No hard colors (so device themes win); justified indented body; headings stay
# whole and never hyphenate; chapters start on a fresh page.
set EPUBCSS (mktemp --suffix=.css)
printf '%s\n' \
  'body{font-family:serif;line-height:1.5;margin:0 5%;text-align:justify;}' \
  'h1,h2,h3,h4{font-weight:bold;line-height:1.2;text-align:left;hyphens:none;-epub-hyphens:none;adobe-hyphenate:none;page-break-after:avoid;}' \
  'h1{text-align:center;font-size:1.8em;margin:1em 0;page-break-before:always;}' \
  'h2{font-size:1.4em;margin:1.6em 0 .6em;}' \
  'h3{font-size:1.15em;margin:1.4em 0 .5em;}' \
  'p{margin:0;text-indent:1.3em;hyphens:auto;-epub-hyphens:auto;orphans:2;widows:2;}' \
  'h1 + p,h2 + p,h3 + p,h4 + p,blockquote p:first-child{text-indent:0;}' \
  'blockquote{margin:1em 1.4em;font-style:italic;}' \
  'ul,ol{margin:1em 0;}' \
  'li{margin:.25em 0;}' \
  'hr{border:0;text-align:center;margin:1.6em 0;}' \
  'hr::after{content:"\2766";}' \
  > $EPUBCSS

echo -n "pdf ... "
pandoc -o $OUT.pdf $IN; and echo "Done"; or echo "FAIL"

echo -n "html ... "
pandoc -s $IN -o $OUT.html \
  --toc --toc-depth=2 \
  --embed-resources \
  --css $HTMLCSS \
  --metadata lang=en-US; and echo "Done"; or echo "FAIL"

echo -n "epub ... "
pandoc -s $IN -o $OUT.epub \
  --toc --toc-depth=2 \
  --split-level=2 \
  --css $EPUBCSS \
  --metadata lang=en-US; and echo "Done"; or echo "FAIL"

command rm -f $HTMLCSS $EPUBCSS

set WWW "../release"
command mkdir -p $WWW
command mv -f $OUT.pdf $OUT.html $OUT.epub $WWW/

set PR "../push_release"
if test -e $PR
  echo "Release"
  pushd $(dirname $PR);
  set -l NAME "./$(basename $PR)";
  command $NAME;
  popd
end;
