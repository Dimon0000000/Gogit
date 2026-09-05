package suggest

// Kind identifies what a suggestion represents.
type Kind string

const (
	KindSubcommand Kind = "subcommand"
	KindOption     Kind = "option"
)

// Suggestion is one value Gogit can insert into the current token.
type Suggestion struct {
	Value       string
	Description string
	Kind        Kind
}

// Context describes the token under the cursor. All offsets are rune indexes.
type Context struct {
	WordsBefore []string
	Token       string
	Prefix      string
	TokenStart  int
	TokenEnd    int
}
