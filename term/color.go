package term

import "github.com/fatih/color"

var (
	GreenHighlight  = color.New(color.FgGreen).SprintFunc()
	RedHighlight    = color.New(color.FgRed).SprintFunc()
	YellowHighlight = color.New(color.FgYellow).SprintFunc()

	Underline = color.New(color.Underline).SprintFunc()

	Highlight = GreenHighlight
)
