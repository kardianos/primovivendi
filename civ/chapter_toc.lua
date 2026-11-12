-- Pandoc Lua filter for the book PDF (build.fish, build_book).
--
-- Chapter headings are written "Chapter N: Title". The printed chapter
-- heading and the running head keep that text. The table of contents
-- gets a number column instead: "N  Title", with wrapped lines aligned
-- under the title (\numberline, laid out by tocloft in the preamble).
-- The PDF bookmark, which would drop the \numberline, reads "N. Title".
-- "(Principle N)" keeps its number on the same line as the word.
-- Other formats and other headings pass through unchanged.

local function tex(s)
  s = s:gsub('[\\{}$&#^_%%~]', function(c) return '\\' .. c end)
  s = s:gsub('Principles? %d', function(m) return (m:gsub(' ', '~')) end)
  return s
end

function Header(h)
  if not FORMAT:match('latex') then
    return nil
  end
  local text = pandoc.utils.stringify(h.content)
  local n, title = text:match('^Chapter (%d+):%s*(.+)$')
  if not n then
    return nil
  end
  local full = tex(text)
  local entry = '\\protect\\numberline{' .. n .. '}' .. tex(title)
  local bookmark = n .. '. ' .. tex(title)
  return pandoc.RawBlock('latex',
    '\\chapter*{' .. full .. '}' ..
    '\\addcontentsline{toc}{chapter}{\\texorpdfstring{' .. entry .. '}{' .. bookmark .. '}}' ..
    '\\markboth{' .. full .. '}{}')
end
