package cfg

import (
	"strings"

	"github.com/rs/xid"
)

type UUIDVarResolver struct{}

func (r *UUIDVarResolver) Resolve(in string) (string, error) {
	return strings.Replace(in, "$UUID", xid.New().String(), -1), nil
}
