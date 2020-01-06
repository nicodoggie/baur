package term

import (
	"fmt"
	"io"
)

const separator = "------------------------------------------------------------------------------"

// PrintSep prints a separator line
func PrintSep(w io.Writer) {
	fmt.Fprintln(w, separator)
}
