package refutil

import (
	"fmt"
	"strings"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/spec"
)

func Append(ref spec.Ref, tks ...string) spec.Ref {
	path := ref.GetURL().Path
	ptr := ref.GetPointer()
	newTks := make([]string, 0, len(tks)+len(ptr.DecodedTokens()))
	for _, tk := range append(ptr.DecodedTokens(), tks...) {
		newTks = append(newTks, jsonpointer.Escape(tk))
	}
	ptrstr := "/" + strings.Join(newTks, "/")
	pptr, err := jsonpointer.New(ptrstr)
	if err != nil {
		panic(fmt.Sprintf("creating json pointer for %s: %v", ptrstr, err))
	}
	return spec.MustCreateRef(path + "#" + pptr.String())
}

func Parent(ref spec.Ref) spec.Ref {
	path := ref.GetURL().Path
	tks := ref.GetPointer().DecodedTokens()
	var newTks []string
	for _, tk := range tks[:len(tks)-1] {
		newTks = append(newTks, jsonpointer.Escape(tk))
	}
	ptrstr := "/" + strings.Join(newTks, "/")
	pptr, err := jsonpointer.New(ptrstr)
	if err != nil {
		panic(fmt.Sprintf("creating json pointer for %s: %v", ptrstr, err))
	}
	return spec.MustCreateRef(path + "#" + pptr.String())
}
