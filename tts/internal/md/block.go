package md

// Kind classifies a structural block of a chapter.
type Kind int

const (
	KindHeading Kind = iota
	KindParagraph
	KindListItem
)

// Block is one structural unit after markdown is parsed for speech.
type Block struct {
	Kind  Kind
	Level int    // heading level 1–6; 0 otherwise
	Text  string // plain text (markdown markers stripped)
}
