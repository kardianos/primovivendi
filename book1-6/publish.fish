#!/usr/bin/fish

# Publication-grade build for book1-6 (Life Under Axioms, working title).
#
# Produces three deliverables under ../release/ (names chosen not to
# collide with book1-4's cf_book.*, book1-5's cfc-book.*, or the older
# po_book.* / the_conservative_frame.*):
#   lua-book.pdf  - print PDF, LaTeX book class, fonts embedded by xelatex
#                   + optional ghostscript pass for print platforms
#   lua-book.epub - reflowable EPUB3 for KDP / Apple Books / Kobo
#   lua-book.html - standalone web edition
#
# Intermediate: book.md (concat of chapters/), work files under ../.out/
#
# concat.go revision modes:
#   go run concat.go        → book.md (new/publish: keep %+, drop %-, drop %%)
#   go run concat.go -old   → book-old.md (old view: keep %-, drop %+)
# This script always runs the new/publish path. Never pass -old here.
# Line markers and HTML writer comments are stripped by concat before pandoc.
#
# Required: go, pandoc, texlive-xetex
# Optional: ghostscript (print PDF pass)

set script_dir (dirname (status --current-filename))
cd $script_dir

# Release builds use new mode only (default). Do not pass -old.
go run concat.go
if test $status -ne 0
  echo "concat failed"
  exit 1
end

set SRC "book"
set DST "lua-book"
set OUT "../.out"
set WWW "../release"
command mkdir -p $OUT $WWW

# ---- book design knobs (match book1-5 publish.fish) ----
set TRIM_W   "6in"
set TRIM_H   "9in"
set INNER    "0.875in"
set OUTER    "0.625in"
set TOP_M    "0.75in"
set BOT_M    "0.75in"
set FONTSIZE "11pt"
set LEADING  "1.08"
set BODYFONT "TeX Gyre Pagella"
set SANSFONT "TeX Gyre Heros"
set MONOFONT "TeX Gyre Cursor"
set PDFX     "on"

# ---- LaTeX preamble fragment (book typography) ----
set HDR (mktemp --suffix=.tex)
printf '%s\n' \
  '\usepackage{titlesec}' \
  '% \nohyph: suppress word-wrap hyphenation in display text (titles, headings,' \
  '% running heads, TOC). Body paragraphs keep hyphenation for clean justification.' \
  '\newcommand{\nohyph}{\hyphenpenalty=10000\exhyphenpenalty=10000}' \
  '\titleformat{\chapter}[display]{\normalfont\huge\bfseries\raggedright\nohyph}{}{0pt}{\Huge}' \
  '\titleformat*{\section}{\normalfont\Large\bfseries\raggedright\nohyph}' \
  '\titleformat*{\subsection}{\normalfont\large\bfseries\raggedright\nohyph}' \
  '\titleformat*{\subsubsection}{\normalfont\normalsize\bfseries\raggedright\nohyph}' \
  '\usepackage{fancyhdr}' \
  '\pagestyle{fancy}' \
  '\fancyhf{}' \
  '\fancyfoot[LE,RO]{\thepage}' \
  '\fancyhead[RE]{{\nohyph\leftmark}}' \
  '\fancyhead[LO]{{\nohyph\rightmark}}' \
  '\renewcommand{\chaptermark}[1]{\markboth{#1}{}}' \
  '\renewcommand{\headrulewidth}{0pt}' \
  '% never strand a single line of a paragraph at the bottom (orphan/club)' \
  '% or top (widow) of a page; force the break to carry a partner line along.' \
  '\clubpenalty=10000' \
  '\widowpenalty=10000' \
  '\displaywidowpenalty=10000' \
  '\AtBeginDocument{\addtocontents{toc}{\protect\nohyph}}' \
  '\PassOptionsToPackage{draft}{hyperref}' \
  > $HDR

# ===== 1. Print PDF (xelatex embeds fonts by default) =====
echo -n "pdf ... "
command pandoc "$SRC.md" \
  -o "$OUT/$DST.pdf" \
  --pdf-engine=xelatex \
  --top-level-division=part \
  -V documentclass=book \
  --include-in-header=$HDR \
  -V geometry:"paperwidth=$TRIM_W" \
  -V geometry:"paperheight=$TRIM_H" \
  -V geometry:"inner=$INNER" \
  -V geometry:"outer=$OUTER" \
  -V geometry:"top=$TOP_M" \
  -V geometry:"bottom=$BOT_M" \
  -V geometry:"twoside" \
  -V fontsize=$FONTSIZE \
  -V linestretch=$LEADING \
  -V mainfont="$BODYFONT" \
  -V sansfont="$SANSFONT" \
  -V monofont="$MONOFONT" \
  -V toc \
  -V toc-depth=1 \
  -V numbersections=true \
  -V colorlinks=false \
  -V linkcolor=black \
  -V urlcolor=black

if test $status -ne 0
  echo "FAIL"
  command rm -f $HDR
  exit 1
end

# Fonts are already embedded by xelatex. This ghostscript pass subsets,
# downsamples images, and flattens to grayscale for print platforms.
if test "$PDFX" = "on"; and type -q gs
  echo -n "print ... "
  command gs \
    -dNOPAUSE -dQUIET -dBATCH \
    -sDEVICE=pdfwrite \
    -sColorConversionStrategy=Gray \
    -dPDFSETTINGS="/printer" \
    -dEmbedAllFonts=true \
    -dSubsetFonts=true \
    -sOutputFile="$OUT/$DST.x.pdf" \
    "$OUT/$DST.pdf"
  if test $status -eq 0
    command mv -f "$OUT/$DST.x.pdf" "$WWW/$DST.pdf"
    echo "Done (print PDF)"
  else
    command mv -f "$OUT/$DST.pdf" "$WWW/$DST.pdf"
    echo "gs failed, fell back to xelatex PDF"
  end
else
  command mv -f "$OUT/$DST.pdf" "$WWW/$DST.pdf"
  echo "Done (xelatex PDF)"
end

# ===== 2. EPUB (reflowable ebook) =====
echo -n "epub ... "
command pandoc "$SRC.md" \
  -o "$OUT/$DST.epub" \
  --toc --toc-depth=1 \
  --split-level=2 \
  --metadata lang=en-US
if test $status -ne 0
  echo "FAIL"
  command rm -f $HDR
  exit 1
end
command mv -f "$OUT/$DST.epub" "$WWW/$DST.epub"
echo "Done"

# ===== 3. HTML (web edition) =====
echo -n "html ... "
command pandoc "$SRC.md" \
  -o "$OUT/$DST.html" \
  --standalone \
  --toc --toc-depth=2 \
  --split-level=2 \
  --metadata lang=en-US
if test $status -ne 0
  echo "FAIL"
  command rm -f $HDR
  exit 1
end
command mv -f "$OUT/$DST.html" "$WWW/$DST.html"
echo "Done"

command rm -f $HDR

echo "Wrote:"
echo "  $WWW/$DST.pdf"
echo "  $WWW/$DST.epub"
echo "  $WWW/$DST.html"

set PR "../push_release"
if test -e $PR
  echo "Release"
  pushd (dirname $PR)
  set -l NAME "./"(basename $PR)
  command $NAME
  popd
end
