#!/usr/bin/env fish
# Speak a Markdown range in a local Piper voice (default: en_US-ryan-high).
#
#   fish civ/tools/readback.fish -file civ/chapters/01_ch01.md -from 17 -to 19
#   fish civ/tools/readback.fish -chapter ch01 -from 17 -to 19 --play
#   fish civ/tools/readback.fish -chapter ch01 -from 17 -to 19 -voice en_US-lessac-medium
#
# Headings, tables, rules, and markup are stripped the way civ/linear prepares
# a chapter, so the voice reads prose. The wav is written under ext-ref/tts/out/.
# --play uses aplay, pw-play, or mpv, whichever is on PATH.

set -l script (realpath (status filename))
set -l repo (dirname (dirname (dirname $script)))
set -l py $repo/ext-ref/tts/venv/bin/python
set -l voice_name en_US-ryan-high
set -l outdir $repo/ext-ref/tts/out

set -l from 1
set -l to 0
set -l play 0
set -l file ""
set -l chapter ""

while set -q argv[1]
    switch $argv[1]
        case -file --file
            set file $argv[2]
            set -e argv[1]
            set -e argv[1]
        case -chapter --chapter
            set chapter $argv[2]
            set -e argv[1]
            set -e argv[1]
        case -from --from
            set from $argv[2]
            set -e argv[1]
            set -e argv[1]
        case -to --to
            set to $argv[2]
            set -e argv[1]
            set -e argv[1]
        case -voice --voice
            set voice_name $argv[2]
            set -e argv[1]
            set -e argv[1]
        case --play -play
            set play 1
            set -e argv[1]
        case -h --help
            echo "usage: fish civ/tools/readback.fish (-file PATH | -chapter ch01) [-from N] [-to M] [-voice NAME] [--play]"
            exit 0
        case '*'
            echo "readback: unknown argument $argv[1]" >&2
            exit 2
    end
end

set -l voice $voice_name
if not string match -q '*.onnx' -- $voice
    set voice $repo/ext-ref/tts/voices/$voice.onnx
else if not string match -q '/*' -- $voice
    if not test -f "$voice"; and test -f "$repo/$voice"
        set voice $repo/$voice
    end
end

if test -n "$file"; and test -n "$chapter"
    echo "readback: pass -file or -chapter, not both" >&2
    exit 2
end
if test -z "$file"; and test -z "$chapter"
    echo "readback: pass -file PATH or -chapter ch01" >&2
    exit 2
end

if test -n "$chapter"
    set -l n (string replace -r '^ch' '' -- $chapter)
    set n (string replace -r '^0+' '' -- $n)
    if test -z "$n"
        set n 0
    end
    if not string match -qr '^[0-9]+$' -- $n
        echo "readback: invalid chapter $chapter" >&2
        exit 2
    end
    set -l nn (printf '%02d' $n)
    set -l matches $repo/civ/chapters/*_ch$nn.md
    if test (count $matches) -eq 0
        echo "readback: no chapter matching ch$nn" >&2
        exit 1
    end
    set file $matches[1]
end
if not string match -q '/*' -- $file
    if not test -f "$file"; and test -f "$repo/$file"
        set file $repo/$file
    end
end
if not test -f $file
    echo "readback: no such file $file" >&2
    exit 1
end
if not test -x $py
    echo "readback: piper is not installed at $py" >&2
    exit 1
end
if not test -f $voice
    echo "readback: voice model missing at $voice" >&2
    exit 1
end

set -l stem (string replace -r '\.md$' '' -- (path basename $file))
mkdir -p $outdir
set -l wav $outdir/$stem-$from-$to.wav
set -l txt (mktemp)
set -l prep (dirname $script)/readback_prose.py

# The slicer mirrors civ/linear/prepare.go so the voice reads prose.
python3 $prep $file $from $to $txt
set -l code $status
if test $code -ne 0
    rm -f $txt
    if test $code -eq 3
        echo "readback: no prose in $file lines $from to $to" >&2
    else
        echo "readback: could not prepare prose" >&2
    end
    exit 1
end

echo "readback: $(wc -w <$txt) words -> $wav"
$py -m piper -m $voice -i $txt -f "$wav" --sentence-silence 0.3
set -l code $status
rm -f $txt
if test $code -ne 0
    echo "readback: piper failed" >&2
    exit $code
end
if not test -s "$wav"
    echo "readback: empty wav $wav" >&2
    exit 1
end
echo "readback: wrote $wav"

if test $play -eq 1
    if command -q aplay
        aplay "$wav"
    else if command -q pw-play
        pw-play "$wav"
    else if command -q mpv
        mpv --no-video "$wav"
    else
        echo "readback: no aplay, pw-play, or mpv on PATH" >&2
        exit 1
    end
end
