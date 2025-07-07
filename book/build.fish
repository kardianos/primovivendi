#!/usr/bin/fish
go run concat.go


set OUT "../.out"
set WWW "../release"

command mkdir -p $OUT

function run -a EXT;
	set -l FN_SRC book.md
	set -l FN_DST "po_book.$EXT"
	set -l FN_OUT "$OUT/$FN_DST"
	set -l FN_WWW "$WWW/$FN_DST"
	echo -n "$FN_DST ... "
	
	set -l MOD "$(stat -c %Y $FN_SRC)";
	set -l FINMOD "0";
	if test -e $FN_WWW
		set -f FINMOD "$(stat -c %Y $FN_WWW)";
	end;
	
	if test $MOD -gt $FINMOD
		if test -e $FN_OUT
			command rm -f "$FN_OUT";
		end;
		command pandoc -s "$FN_SRC" -o "$FN_OUT" --split-level=2;
		command mv -f "$FN_OUT" "$FN_WWW";
		echo "Done"
	else
		echo "Skip"
	end;
end;

run pdf;
run html;
run epub;
