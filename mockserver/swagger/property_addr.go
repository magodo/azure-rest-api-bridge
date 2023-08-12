package swagger

import (
	"fmt"
	"strings"

	"github.com/go-openapi/jsonpointer"
)

type PropertyAddr []PropertyAddrStep

func (addr PropertyAddr) Copy() PropertyAddr {
	naddr := make(PropertyAddr, len(addr))
	copy(naddr, addr)
	return naddr
}

type PropertyAddrStep struct {
	Type  PropertyAddrStepType
	Value string
}

var RootAddr = PropertyAddr{}

type PropertyAddrStepType int

const (
	PropertyAddrStepTypeProp PropertyAddrStepType = iota
	PropertyAddrStepTypeIndex
	PropertyAddrStepTypeVariant
)

func (addr PropertyAddr) String() string {
	var addrs []string
	for _, step := range addr {
		switch step.Type {
		case PropertyAddrStepTypeProp:
			addrs = append(addrs, step.Value)
		case PropertyAddrStepTypeIndex:
			addrs = append(addrs, "*")
		case PropertyAddrStepTypeVariant:
			addrs = append(addrs, "{"+step.Value+"}")
		default:
			panic(fmt.Sprintf("unknown step type: %d", step.Type))
		}
	}
	return strings.Join(addrs, ".")
}

func (addr PropertyAddr) Equal(oaddr PropertyAddr) bool {
	if len(addr) != len(oaddr) {
		return false
	}
	for i := range addr {
		seg1, seg2 := addr[i], oaddr[i]
		if seg1.Type != seg2.Type || seg1.Value != seg2.Value {
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
		case PropertyAddrStepTypeVariant:
			continue
		default:
			panic(fmt.Sprintf("unknown step type: %d", step.Type))
		}
	}
	ptrstr := "/" + strings.Join(tks, "/")
	return jsonpointer.New(ptrstr)
}

func ParseAddr(input string) PropertyAddr {
	if input == "" {
		return RootAddr
	}
	var addr PropertyAddr
	for _, part := range strings.Split(input, ".") {
		var step PropertyAddrStep
		if part == "*" {
			step = PropertyAddrStep{Type: PropertyAddrStepTypeIndex}
		} else if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			step = PropertyAddrStep{Type: PropertyAddrStepTypeVariant, Value: strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")}
		} else {
			step = PropertyAddrStep{Type: PropertyAddrStepTypeProp, Value: part}
		}
		addr = append(addr, step)
	}
	return addr
}
