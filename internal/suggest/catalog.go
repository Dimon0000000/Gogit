package suggest

var gitSubcommands = []Suggestion{
	{
		Value:       "branch",
		Description: "List, create, or delete branches.",
		Kind:        KindSubcommand,
	},
}

var gitOptions = map[string][]Suggestion{
	"branch": {
		{
			Value:       "--show-current",
			Description: "Print the name of the current branch.",
			Kind:        KindOption,
		},
		{
			Value:       "--merged",
			Description: "List branches already merged into the specified commit.",
			Kind:        KindOption,
		},
		{
			Value:       "--no-merged",
			Description: "List branches that have not been merged.",
			Kind:        KindOption,
		},
		{
			Value:       "--delete",
			Description: "Delete a branch.",
			Kind:        KindOption,
		},
	},
}
