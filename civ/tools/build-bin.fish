#!/usr/bin/env fish
# Build all civ helper binaries into civ/bin/ (gitignored).
# Run once, or re-run after editing tool code. Does not build or publish the book.

set -l script (realpath (status filename))
set -l civ (dirname (dirname $script))
set -l bin $civ/bin

mkdir -p $bin

for tool in lint linear ngram review compose parser
    echo "building civ/bin/$tool"
    pushd $civ >/dev/null
    env GOEXPERIMENT=simd go build -o $bin/$tool ./$tool
    set -l code $status
    popd >/dev/null
    if test $code -ne 0
        exit $code
    end
end

ln -sf ../tools/readback.fish $bin/readback
echo "ready: $bin/{lint,linear,ngram,review,compose,parser,readback}"
