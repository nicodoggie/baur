package cfg

import "strings"

// StrReplacementResolver replaces the string Old with New.
type StrReplacementResolver struct {
	Old string
	New string
}

func (s *StrReplacementResolver) Resolve(in string) (string, error) {
	return strings.Replace(in, s.Old, s.New, -1), nil
}
