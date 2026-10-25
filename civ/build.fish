#!/usr/bin/fish

# Build each stem in STEMS as pdf, html, and epub.
# Same pipeline as book1-4/essay_build.fish.
# build_stem appends .md. It rebuilds when that file is newer than a
# product, or when a product is missing.
#   <stem>.pdf  - print, via LaTeX, 6in x 9in, TeX Gyre Pagella,
#                 then an optional ghostscript print pass
#   <stem>.html - single-file web edition, book typography embedded
#   <stem>.epub - reflowable ebook, split at the top sections
#
# Writes products in this directory.
#
# The markdown has no YAML title. The first H1 is the title. Lines above
# that H1 are a generation note, so they are left out. The PDF and the
# EPUB shift the remaining headings up one level. In the PDF, ## becomes
# a section and #### stays a displayed subsubsection. Article class would
# otherwise set #### as a run-in paragraph. In the EPUB, ## becomes the
# chapter (h1) so the book splits on the three top sections and does not
# add an empty title chapter. HTML keeps the markdown levels, so ## is an
# h2 under the title, matching the essay stylesheet.
#
# Required: pandoc, texlive-xetex
# Optional: ghostscript (print PDF pass)

set script_dir (dirname (status --current-filename))
cd $script_dir

# Basenames in this directory. build_stem appends .md.
set STEMS ccp_argon ccp_derive_argon ccp_eng_argon

# ---- print PDF (xelatex): book typography, essay layout ----
# Same fonts and heading treatment as book1-4/essay_build.fish, as an
# article on a 6in x 9in page.
set TRIM_W   "6in"
set TRIM_H   "9in"
set MARGIN_X "0.6in"
set MARGIN_Y "0.7in"
set FONTSIZE "11pt"
set LEADING  "1.05"
set BODYFONT "TeX Gyre Pagella"
set SANSFONT "TeX Gyre Heros"
set MONOFONT "TeX Gyre Cursor"
set PDFX     "on"

function build_stem --argument-names stem
  set -l IN "$stem.md"
  set -l OUT $stem

  if not test -f $IN
    echo "missing: $IN"
    exit 1
  end

  # Rebuild when any product is missing or older than the markdown.
  set -l rebuild 0
  for ext in pdf html epub
    if not test -f "$OUT.$ext"; or test "$IN" -nt "$OUT.$ext"
      set rebuild 1
      break
    end
  end
  if test $rebuild -eq 0
    echo "$stem up to date"
    return 0
  end

  echo "build $stem"
  set TITLE (string trim (string replace -r '^# ' '' (grep -m1 -e '^# ' $IN)))
  if test -z "$TITLE"
    echo "no title: $IN has no H1"
    exit 1
  end

  # Body starts after the H1. --shift-heading-level-by=-1 then maps
  # ## / ### / #### onto section / subsection / subsubsection.
  set BODY (mktemp --suffix=.md)
  awk 'BEGIN {seen=0} /^# / && !seen {seen=1; next} seen {print}' $IN > $BODY
  if test -z (grep -m1 -e '^##' $BODY)
    echo "no sections after the title in $IN"
    command rm -f $BODY
    exit 1
  end

  # LaTeX preamble: match the essay heading style (raggedright, never hyphenate
  # display text), centered page number in the footer, no widows/orphans. The
  # title runs centered in the header; the title page uses the plain pagestyle
  # (via \maketitle) so the header appears on every page but the first.
  # PassOptionsToPackage{draft}{hyperref}: strip link annotations so gs can
  # reprocess the PDF in the print pass below.
  set PDFHDR (mktemp --suffix=.tex)
  printf '%s\n' \
    '\usepackage{titlesec}' \
    '\newcommand{\nohyph}{\hyphenpenalty=10000\exhyphenpenalty=10000}' \
    '\titleformat*{\section}{\normalfont\Large\bfseries\raggedright\nohyph}' \
    '\titleformat*{\subsection}{\normalfont\large\bfseries\raggedright\nohyph}' \
    '\titleformat*{\subsubsection}{\normalfont\normalsize\bfseries\raggedright\nohyph}' \
    '\usepackage{fancyhdr}' \
    '\pagestyle{fancy}' \
    '\fancyhf{}' \
    '\fancyfoot[C]{\thepage}' \
    '\fancyhead[C]{{\small\itshape\nohyph '$TITLE'}}' \
    '\renewcommand{\headrulewidth}{0pt}' \
    '\clubpenalty=10000' \
    '\widowpenalty=10000' \
    '\displaywidowpenalty=10000' \
    '\PassOptionsToPackage{draft}{hyperref}' \
    > $PDFHDR

  echo -n "pdf ... "
  command pandoc $BODY \
    -o $OUT.pdf \
    --shift-heading-level-by=-1 \
    --pdf-engine=xelatex \
    --include-in-header=$PDFHDR \
    -M title="$TITLE" \
    -V geometry:"paperwidth=$TRIM_W" \
    -V geometry:"paperheight=$TRIM_H" \
    -V geometry:"left=$MARGIN_X" \
    -V geometry:"right=$MARGIN_X" \
    -V geometry:"top=$MARGIN_Y" \
    -V geometry:"bottom=$MARGIN_Y" \
    -V fontsize=$FONTSIZE \
    -V linestretch=$LEADING \
    -V mainfont="$BODYFONT" \
    -V sansfont="$SANSFONT" \
    -V monofont="$MONOFONT" \
    -V colorlinks=false \
    -V linkcolor=black \
    -V urlcolor=black
  set -l pdf_status $status
  command rm -f $PDFHDR
  if test $pdf_status -ne 0
    echo "FAIL"
    command rm -f $BODY
    exit 1
  end
  echo "Done"

  # Fonts are already embedded by xelatex. This ghostscript pass subsets,
  # downsamples images, and flattens to grayscale for print platforms, matching
  # book1-4/essay_build.fish. Set PDFX=off to skip. Rewrites $OUT.pdf in place.
  if test "$PDFX" = "on"; and type -q gs
    echo -n "print ... "
    command gs \
      -dNOPAUSE -dQUIET -dBATCH \
      -sDEVICE=pdfwrite \
      -sColorConversionStrategy=Gray \
      -dPDFSETTINGS="/printer" \
      -dEmbedAllFonts=true \
      -dSubsetFonts=true \
      -sOutputFile="$OUT.x.pdf" \
      "$OUT.pdf"
    if test $status -eq 0
      command mv -f "$OUT.x.pdf" "$OUT.pdf"
      echo "Done (print PDF)"
    else
      command rm -f "$OUT.x.pdf"
      echo "gs failed, kept xelatex PDF"
    end
  end

  # ---- web (HTML) stylesheet: a centered reading column that reads like a page ----
  # Justified, first-line-indented body; headings never hyphenate; styled title
  # block + table of contents; respects the reader's light/dark preference.
  # h4 is the gloss level (####), which the essay stylesheet does not size.
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
    'h4{margin:1.4em 0 .4em;font-size:1.05rem;}' \
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
  # whole and never hyphenate. The file splits at h1 (the former ## sections),
  # so each top section opens on its own page.
  set EPUBCSS (mktemp --suffix=.css)
  printf '%s\n' \
    'body{font-family:serif;line-height:1.5;margin:0 5%;text-align:justify;}' \
    'h1,h2,h3,h4{font-weight:bold;line-height:1.2;text-align:left;hyphens:none;-epub-hyphens:none;adobe-hyphenate:none;page-break-after:avoid;}' \
    'h1{text-align:center;font-size:1.8em;margin:1em 0;page-break-before:always;}' \
    'h2{font-size:1.4em;margin:1.6em 0 .6em;}' \
    'h3{font-size:1.15em;margin:1.4em 0 .5em;}' \
    'h4{font-size:1.05em;margin:1.2em 0 .4em;}' \
    'p{margin:0;text-indent:1.3em;hyphens:auto;-epub-hyphens:auto;orphans:2;widows:2;}' \
    'h1 + p,h2 + p,h3 + p,h4 + p,blockquote p:first-child{text-indent:0;}' \
    'blockquote{margin:1em 1.4em;font-style:italic;}' \
    'ul,ol{margin:1em 0;}' \
    'li{margin:.25em 0;}' \
    'hr{border:0;text-align:center;margin:1.6em 0;}' \
    'hr::after{content:"\2766";}' \
    > $EPUBCSS

  echo -n "html ... "
  command pandoc -s $BODY \
    -o $OUT.html \
    --toc --toc-depth=2 \
    --embed-resources \
    --css $HTMLCSS \
    -M title="$TITLE" \
    --metadata lang=en-US
  if test $status -ne 0
    echo "FAIL"
    command rm -f $HTMLCSS $EPUBCSS $BODY
    exit 1
  end
  echo "Done"

  echo -n "epub ... "
  command pandoc -s $BODY \
    -o $OUT.epub \
    --shift-heading-level-by=-1 \
    --toc --toc-depth=2 \
    --split-level=1 \
    --css $EPUBCSS \
    -M title="$TITLE" \
    --metadata lang=en-US
  if test $status -ne 0
    echo "FAIL"
    command rm -f $HTMLCSS $EPUBCSS $BODY
    exit 1
  end
  echo "Done"

  command rm -f $HTMLCSS $EPUBCSS $BODY

end

for stem in $STEMS
  build_stem $stem
end

function build_book --argument-names stem chdir
  set -l IN "$stem.md"
  set -l OUT $stem

  # Assemble chapters/*.md into $IN when $IN is missing or older than any chapter file.
  if test -d "$chdir"
    set -l ch_files (command ls "$chdir"/*.md 2>/dev/null | sort)
    if test (count $ch_files) -gt 0
      set -l need_concat 0
      if not test -f "$IN"
        set need_concat 1
      else
        for cf in $ch_files
          if test "$cf" -nt "$IN"
            set need_concat 1
            break
          end
        end
      end
      if test $need_concat -eq 1
        echo "assemble $IN from $chdir/"
        python3 -c '
import sys
out_path = sys.argv[1]
files = sys.argv[2:]
chunks = []
for p in files:
    with open(p) as f:
        chunks.append(f.read().strip())
with open(out_path, "w") as f:
    f.write("\n\n".join(chunks) + "\n")
' "$IN" $ch_files
      end
    end
  end

  if not test -f $IN
    echo "missing: $IN"
    exit 1
  end

  # Rebuild when any product is missing or older than the markdown or build.fish.
  set -l rebuild 0
  for ext in pdf html epub
    if not test -f "$OUT.$ext"; or test "$IN" -nt "$OUT.$ext"; or test "build.fish" -nt "$OUT.$ext"
      set rebuild 1
      break
    end
  end
  if test $rebuild -eq 0
    echo "$stem up to date"
    return 0
  end

  echo "build book $stem"
  set TITLE (string trim (string replace -r '^# ' '' (grep -m1 -e '^# ' $IN)))
  if test -z "$TITLE"
    echo "no title: $IN has no H1"
    exit 1
  end
  set SHORT_TITLE (string trim (string replace -r ':.*$' '' "$TITLE"))
  set SHORT_TITLE_TEX (string replace -a "'" '\textquotesingle ' "$SHORT_TITLE")

  set BODY (mktemp --suffix=.md)
  awk 'BEGIN {seen=0} /^# / && !seen {seen=1; next} seen {print}' $IN > $BODY

  # LaTeX book preamble:
  # - Each chapter starts on a new page (\chapter in book class with openany).
  # - TOC lists chapters (--toc --toc-depth=1).
  # - Two-sided even/odd headers and footers:
  #     Even page header (LE): Book Title
  #     Odd page header  (RO): Chapter Name (\leftmark)
  #     Even/Odd footer  (LE,RO): Page number on outer corners
  set PDFHDR (mktemp --suffix=.tex)
  printf '%s\n' \
    '\usepackage{titlesec}' \
    '\usepackage{emptypage}' \
    '\newcommand{\nohyph}{\hyphenpenalty=10000\exhyphenpenalty=10000}' \
    '\titleformat{\chapter}[display]{\normalfont\huge\bfseries\raggedright\nohyph}{}{0pt}{\Huge}' \
    '\titleformat*{\section}{\normalfont\Large\bfseries\raggedright\nohyph}' \
    '\titleformat*{\subsection}{\normalfont\large\bfseries\raggedright\nohyph}' \
    '\titleformat*{\subsubsection}{\normalfont\normalsize\bfseries\raggedright\nohyph}' \
    '\usepackage{fancyhdr}' \
    '\setlength{\headheight}{28pt}' \
    '\pagestyle{fancy}' \
    '\fancyhf{}' \
    '\fancyfoot[LE,RO]{\thepage}' \
    '\fancyhead[LE]{{\small\itshape\nohyph '$SHORT_TITLE_TEX'}}' \
    '\fancyhead[RO]{\parbox[b]{\textwidth}{\raggedleft\small\itshape\nohyph\leftmark}}' \
    '\fancypagestyle{plain}{\fancyhf{}\fancyfoot[LE,RO]{\thepage}\renewcommand{\headrulewidth}{0pt}}' \
    '\renewcommand{\chaptermark}[1]{\markboth{#1}{}}' \
    '\renewcommand{\sectionmark}[1]{}' \
    '\renewcommand{\headrulewidth}{0pt}' \
    '\makeatletter\renewcommand\tableofcontents{\chapter*{\contentsname\@mkboth{\contentsname}{\contentsname}}\@starttoc{toc}}\makeatother' \
    '\clubpenalty=10000' \
    '\widowpenalty=10000' \
    '\displaywidowpenalty=10000' \
    '\AtBeginDocument{\addtocontents{toc}{\protect\nohyph}}' \
    '\PassOptionsToPackage{draft}{hyperref}' \
    > $PDFHDR

  echo -n "pdf ... "
  command pandoc $BODY \
    -o $OUT.pdf \
    --shift-heading-level-by=-1 \
    --top-level-division=chapter \
    --pdf-engine=xelatex \
    -V documentclass=book \
    -V classoption=openany \
    --include-in-header=$PDFHDR \
    -M title="$TITLE" \
    -V geometry:"paperwidth=$TRIM_W" \
    -V geometry:"paperheight=$TRIM_H" \
    -V geometry:"inner=0.875in" \
    -V geometry:"outer=0.625in" \
    -V geometry:"top=0.85in" \
    -V geometry:"bottom=0.75in" \
    -V geometry:"twoside" \
    -V fontsize=$FONTSIZE \
    -V linestretch=1.08 \
    -V mainfont="$BODYFONT" \
    -V sansfont="$SANSFONT" \
    -V monofont="$MONOFONT" \
    --toc --toc-depth=1 \
    -V colorlinks=false \
    -V linkcolor=black \
    -V urlcolor=black
  set -l pdf_status $status
  command rm -f $PDFHDR
  if test $pdf_status -ne 0
    echo "FAIL"
    command rm -f $BODY
    exit 1
  end
  echo "Done"

  if test "$PDFX" = "on"; and type -q gs
    echo -n "print ... "
    command gs \
      -dNOPAUSE -dQUIET -dBATCH \
      -sDEVICE=pdfwrite \
      -sColorConversionStrategy=Gray \
      -dPDFSETTINGS="/printer" \
      -dEmbedAllFonts=true \
      -dSubsetFonts=true \
      -sOutputFile="$OUT.x.pdf" \
      "$OUT.pdf"
    if test $status -eq 0
      command mv -f "$OUT.x.pdf" "$OUT.pdf"
      echo "Done (print PDF)"
    else
      command rm -f "$OUT.x.pdf"
      echo "gs failed, kept xelatex PDF"
    end
  end

  set HTMLCSS (mktemp --suffix=.css)
  printf '%s\n' \
    ':root{--measure:34em;--ink:#1a1a1a;--paper:#fdfdfb;--muted:#6a6a6a;--link:#0b3d91;}' \
    '*{box-sizing:border-box;}' \
    'html{font-size:112.5%;}' \
    'body{font-family:"Iowan Old Style","Palatino Linotype","Book Antiqua",Palatino,Georgia,Cambria,serif;line-height:1.6;color:var(--ink);background:var(--paper);max-width:var(--measure);margin:0 auto;padding:3.5rem 1.25rem 6rem;text-rendering:optimizeLegibility;-webkit-font-smoothing:antialiased;}' \
    'p{margin:0;text-align:justify;hyphens:auto;}' \
    'p + p{text-indent:1.4em;}' \
    'h1,h2,h3,h4{font-weight:700;line-height:1.25;hyphens:none;}' \
    'h2{margin:2.4em 0 .6em;font-size:1.5rem;page-break-before:always;}' \
    'h3{margin:1.8em 0 .5em;font-size:1.2rem;}' \
    'h4{margin:1.4em 0 .4em;font-size:1.05rem;}' \
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

  set EPUBCSS (mktemp --suffix=.css)
  printf '%s\n' \
    'body{font-family:serif;line-height:1.5;margin:0 5%;text-align:justify;}' \
    'h1,h2,h3,h4{font-weight:bold;line-height:1.2;text-align:left;hyphens:none;-epub-hyphens:none;adobe-hyphenate:none;page-break-after:avoid;}' \
    'h1{text-align:center;font-size:1.8em;margin:1em 0;page-break-before:always;}' \
    'h2{font-size:1.4em;margin:1.6em 0 .6em;}' \
    'h3{font-size:1.15em;margin:1.4em 0 .5em;}' \
    'h4{font-size:1.05em;margin:1.2em 0 .4em;}' \
    'p{margin:0;text-indent:1.3em;hyphens:auto;-epub-hyphens:auto;orphans:2;widows:2;}' \
    'h1 + p,h2 + p,h3 + p,h4 + p,blockquote p:first-child{text-indent:0;}' \
    'blockquote{margin:1em 1.4em;font-style:italic;}' \
    'ul,ol{margin:1em 0;}' \
    'li{margin:.25em 0;}' \
    'hr{border:0;text-align:center;margin:1.6em 0;}' \
    'hr::after{content:"\2766";}' \
    > $EPUBCSS

  echo -n "html ... "
  command pandoc -s $BODY \
    -o $OUT.html \
    --toc --toc-depth=1 \
    --embed-resources \
    --css $HTMLCSS \
    -M title="$TITLE" \
    --metadata lang=en-US
  if test $status -ne 0
    echo "FAIL"
    command rm -f $HTMLCSS $EPUBCSS $BODY
    exit 1
  end
  echo "Done"

  echo -n "epub ... "
  command pandoc -s $BODY \
    -o $OUT.epub \
    --shift-heading-level-by=-1 \
    --toc --toc-depth=1 \
    --split-level=1 \
    --css $EPUBCSS \
    -M title="$TITLE" \
    --metadata lang=en-US
  if test $status -ne 0
    echo "FAIL"
    command rm -f $HTMLCSS $EPUBCSS $BODY
    exit 1
  end
  echo "Done"

  command rm -f $HTMLCSS $EPUBCSS $BODY
end

build_book ccp_book_argon chapters
