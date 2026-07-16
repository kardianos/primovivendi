#!/usr/bin/env fish
set -l root (dirname (status -f))/..
exec $root/.venv/bin/python $root/scripts/harness.py $argv
