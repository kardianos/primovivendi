#!/usr/bin/fish
set script_dir (dirname (status --current-filename));
cd $script_dir;

set OUT "../.out"
set WWW "../release"

command mkdir -p $OUT

function run -a NAME -a VER -a EXT;
	set -l FN_SRC {$NAME}_{$VER}.md
	set -l FN_DST "$NAME.$EXT"
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
		command pandoc -s "$FN_SRC" -o "$FN_OUT";
		command mv -f "$FN_OUT" "$FN_WWW";
		echo "Done"
	else
		echo "Skip"
	end;
end;

run po_summary v3 pdf;
run po_summary v3 html;

run po_outline v3 pdf;
run po_outline v3 html;


set PR "../push_release"
if test -e $PR
	echo "Release"
	pushd $(dirname $PR);
	set -l NAME "./$(basename $PR)";
	command $NAME;
	popd
end;
