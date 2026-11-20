#!/usr/bin/env python3
"""Slice a Markdown file and strip it to prose, matching civ/linear/prepare.go.

Usage: readback_prose.py FILE FROM TO DEST
FROM and TO are 1-based inclusive line numbers. TO of 0 means the end.
Exits 3 when the slice has no prose.
"""

import re
import sys

path, from_s, to_s, dest = sys.argv[1:]
start, end = int(from_s), int(to_s)
with open(path, encoding="utf-8") as f:
    lines = f.read().split("\n")
if start < 1:
    start = 1
if end <= 0 or end > len(lines):
    end = len(lines)
lines = lines[start - 1 : end]

re_link = re.compile(r"\[([^\]]*)\]\([^)]*\)")
re_foot = re.compile(r"\[\^[^\]]*\]")
re_html = re.compile(r"<[^>]+>")
re_list = re.compile(r"^\s*(?:[*+-]|\d+\.)\s+")
re_spaces = re.compile(r"[ \t]+")
re_italic_q = re.compile(r"\*+[\"“][^\"”\n]*[\"”]\*+")
re_long_q = re.compile(r"[\"“][^\"”\n]{80,}[\"”]")


def unquote(t):
    q = False
    while t.startswith(">"):
        t = t[1:].strip()
        q = True
    return t, q


def whole_quotation(t):
    t = re_long_q.sub("", re_italic_q.sub("", t))
    return not any(ch.isalpha() for ch in t)


paras = []
cur = None
fence = False
quote = False


def flush():
    global cur
    if cur is not None:
        paras.append(cur)
        cur = None


for line in lines:
    t = line.strip()
    if t.startswith("```"):
        fence = not fence
        flush()
        continue
    if fence:
        continue
    t, q = unquote(t)
    if q != quote:
        flush()
        quote = q
    if q:
        line = t
    if (
        t == ""
        or t.startswith("#")
        or t.startswith("|")
        or t.startswith("---")
        or t.startswith("<!--")
        or t.startswith(":::")
    ):
        flush()
        continue
    if re_list.match(line):
        flush()
        line = re_list.sub("", line)
    if cur is None:
        cur = {"quote": q, "text": []}
    cur["text"].append(line.strip())
flush()

rep = str.maketrans({"*": "", "`": "", "\\": ""})
out = []
for p in paras:
    t = " ".join(p["text"])
    if p["quote"] and whole_quotation(t):
        continue
    t = re_link.sub(r"\1", t)
    t = re_foot.sub("", t)
    t = re_html.sub(" ", t)
    t = t.translate(rep)
    t = re_spaces.sub(" ", t).strip()
    if t:
        out.append(t)
prose = "\n\n".join(out).strip() + ("\n" if out else "")
with open(dest, "w", encoding="utf-8") as f:
    f.write(prose)
if not prose.strip():
    sys.exit(3)
