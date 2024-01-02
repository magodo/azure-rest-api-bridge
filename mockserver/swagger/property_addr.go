package swagger

import (
	"errors"
	"fmt"
	"strings"
	"text/scanner"

	"github.com/go-openapi/jsonpointer"
)

const (
	delimRune        rune = '/'
	variantOpenRune       = '{'
	variantCloseRune      = '}'
	escapeRune            = '\\'
	indexRune             = '*'
)

type PropertyAddr []PropertyAddrStep

func (addr PropertyAddr) Copy() PropertyAddr {
	naddr := make(PropertyAddr, len(addr))
	copy(naddr, addr)
	return naddr
}

type PropertyAddrStep struct {
	Type    PropertyAddrStepType
	Value   string
	Variant string
}

func (step PropertyAddrStep) String() string {
	var v string
	switch step.Type {
	case PropertyAddrStepTypeIndex:
		v = string(indexRune)
	case PropertyAddrStepTypeProp:
		v = escapePropValue(step.Value)
	default:
		panic(fmt.Sprintf("unknown step type: %d", step.Type))
	}

	if step.Variant != "" {
		v += "{" + escapeVariantValue(step.Variant) + "}"
	}
	return v
}

func escapePropValue(v string) string {
	v = strings.ReplaceAll(v, string(escapeRune), strings.Repeat(string(escapeRune), 2))
	v = strings.ReplaceAll(v, string(variantOpenRune), string(escapeRune)+string(variantCloseRune))
	v = strings.ReplaceAll(v, string(delimRune), string(escapeRune)+string(delimRune))
	return v
}

func escapeVariantValue(v string) string {
	v = strings.ReplaceAll(v, string(escapeRune), strings.Repeat(string(escapeRune), 2))
	v = strings.ReplaceAll(v, string(variantCloseRune), string(escapeRune)+string(variantCloseRune))
	return v
}

var RootAddr = PropertyAddr{}

type PropertyAddrStepType int

const (
	PropertyAddrStepTypeProp PropertyAddrStepType = iota
	PropertyAddrStepTypeIndex
)

func (addr PropertyAddr) String() string {
	var addrs []string
	for _, step := range addr {
		addrs = append(addrs, step.String())
	}
	return strings.Join(addrs, string(delimRune))
}

func (addr PropertyAddr) Equal(oaddr PropertyAddr) bool {
	if len(addr) != len(oaddr) {
		return false
	}
	for i := range addr {
		seg1, seg2 := addr[i], oaddr[i]
		if seg1 != seg2 {
			return false
		}
	}
	return true
}

func (addr PropertyAddr) ToPointer() (jsonpointer.Pointer, error) {
	var tks []string
	for _, step := range addr {
		switch step.Type {
		case PropertyAddrStepTypeIndex:
			tks = append(tks, "0")
		case PropertyAddrStepTypeProp:
			tks = append(tks, step.Value)
		default:
			panic(fmt.Sprintf("unknown step type: %d", step.Type))
		}
	}
	ptrstr := "/" + strings.Join(tks, "/")
	return jsonpointer.New(ptrstr)
}

func ParseAddr(input string) (*PropertyAddr, error) {
	if input == "" {
		return &RootAddr, nil
	}

	var (
		addr        PropertyAddr
		s           scanner.Scanner
		inVariant   bool
		stepValue   string
		stepVariant string
	)

	s.Init(strings.NewReader(input))
	var err error
	s.Error = func(s *scanner.Scanner, msg string) {
		err = errors.Join(err, errors.New(msg))
	}
	for tk := s.Next(); ; tk = s.Next() {
		switch tk {
		case delimRune, scanner.EOF:
			if inVariant {
				stepVariant += string(tk)
			} else {
				if stepValue == "" && stepVariant == "" {
					return nil, fmt.Errorf("both step value and step variant is empty")
				}
				if stepValue == string(indexRune) {
					addr = append(addr, PropertyAddrStep{Type: PropertyAddrStepTypeIndex, Variant: stepVariant})
				} else {
					addr = append(addr, PropertyAddrStep{Type: PropertyAddrStepTypeProp, Value: stepValue, Variant: stepVariant})
				}
				stepValue = ""
				stepVariant = ""
			}
			if tk == scanner.EOF {
				return &addr, err
			}
		case escapeRune:
			switch pk := s.Peek(); pk {
			case delimRune, escapeRune, variantOpenRune, variantCloseRune:
				if inVariant {
					stepVariant += string(pk)
				} else {
					stepValue += string(pk)
				}
				s.Next()
			default:
				return nil, fmt.Errorf("invalid escape %q", `\`+string(pk))
			}
		case variantOpenRune:
			if inVariant {
				stepVariant += string(tk)
			}
			inVariant = true
		case variantCloseRune:
			if !inVariant {
				stepValue += string(tk)
			} else {
				if pk := s.Peek(); pk != delimRune && pk != scanner.EOF {
					return nil, fmt.Errorf(`variant value ends with additional tokens`)
				}
				if stepVariant == "" {
					return nil, fmt.Errorf(`empty variant value`)
				}
				inVariant = false
			}
		default:
			if inVariant {
				stepVariant += string(tk)
			} else {
				stepValue += string(tk)
			}
		}
	}
}

func MustParseAddr(input string) PropertyAddr {
	addr, err := ParseAddr(input)
	if err != nil {
		panic(err)
	}
	return *addr
}
