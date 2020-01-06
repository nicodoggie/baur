package util

import (
	"github.com/fatih/color"
	"github.com/simplesurance/baur"
)

var (
	GreenHighlight  = color.New(color.FgGreen).SprintFunc()
	RedHighlight    = color.New(color.FgRed).SprintFunc()
	YellowHighlight = color.New(color.FgYellow).SprintFunc()
	Underline       = color.New(color.Underline).SprintFunc()
	// highlight is a function that highlights parts of strings in the cli output
	Highlight = GreenHighlight
)

func ColoredBuildStatus(status baur.BuildStatus) string {
	switch status {
	case baur.BuildStatusInputsUndefined:
		return YellowHighlight(status.String())
	case baur.BuildStatusBuildCommandUndefined:
		return YellowHighlight(status.String())
	case baur.BuildStatusExist:
		return GreenHighlight(status.String())
	case baur.BuildStatusPending:
		return RedHighlight(status.String())
	default:
		return status.String()
	}
}
