package refutil

import (
	"strings"

	"github.com/go-openapi/jsonreference"
)

func Last(ref jsonreference.Ref) string {
	if ref.String() == "" {
		return ""
	}
	segs := strings.Split(ref.GetURL().Fragment, "/")
	return segs[len(segs)-1]
}
