package swagger

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/rickb777/date/period"
)

// We use following command to find all the used formats in Swagger (mgmt plane):
//
// Command:
// $ find specification/*/resource-manager -not -path "*/examples/*" -type f -name "*.json" | xargs -I %  bash -c "jq -r '.. | select (.format? != null) | .format |  select(type == \"string\")' < %"  | sort | uniq
//
// This is what it prints:
// arm-id
// base64url
// binary
// byte
// date
// date-time
// date-time-rfc1123
// decimal
// double
// duration
// email
// file
// float
// int32
// int64
// password
// time
// unixtime
// uri
// url
// uuid

type Rnd struct {
	rawString  string
	rawInteger int64
	rawNumber  float64
	// We explicitly not include boolean as it has only two possible values

	time time.Time
}

type RndOption struct {
	InitString  string
	InitInteger int64
	InitNumber  float64

	InitTime time.Time
}

func NewRnd(opt *RndOption) Rnd {
	if opt == nil {
		opt = &RndOption{
			InitString:  "a",
			InitInteger: 0,
			InitNumber:  0.5,
			InitTime:    time.Now(),
		}
	}
	return Rnd{
		rawString:  opt.InitString,
		rawInteger: opt.InitInteger,
		rawNumber:  opt.InitNumber,
		time:       opt.InitTime,
	}
}

func (rnd Rnd) genString(format string) string {
	switch format {
	case "arm-id":
		return "/subscriptions/00000000-0000-0000-000000000000/resourceGroups/" + rnd.rawString
	case "base64url", "byte":
		return base64.StdEncoding.EncodeToString([]byte(rnd.rawString))
	case "binary":
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, rnd.rawInteger)
		if err != nil {
			panic(fmt.Sprintf("binary.Write failed: %v", err))
		}
		return buf.String()
	case "date":
		return rnd.time.Format("2006-01-02")
	case "date-time":
		return rnd.time.Format(time.RFC3339)
	case "date-time-rfc1123":
		return rnd.time.Format(time.RFC1123)
	case "duration":
		p, _ := period.NewOf(time.Duration(rnd.rawInteger) * time.Hour)
		return p.String()
	case "email":
		return rnd.rawString + "@foo.com"
	case "file", "password":
		return rnd.rawString
	case "time":
		return rnd.time.Format("15:04:05")
	case "uri", "url":
		return "https://" + rnd.rawString + ".com"
	case "uuid":
		panic("TODO: support uuid format")
	default:
		return rnd.rawString
	}
}

func (rnd *Rnd) NextString(format string) string {
	switch format {
	case "arm-id":
		rnd.updateRawString()
	case "base64url", "byte":
		rnd.updateRawString()
	case "binary":
		rnd.updateRawInteger()
	case "date":
		rnd.nextRawTime(time.Hour * time.Duration(24))
	case "date-time":
		rnd.nextRawTime(time.Hour)
	case "date-time-rfc1123":
		rnd.nextRawTime(time.Hour)
	case "duration":
		rnd.updateRawInteger()
	case "email":
		rnd.updateRawString()
	case "file", "password":
		rnd.updateRawString()
	case "time":
		rnd.nextRawTime(time.Hour)
	case "uri", "url":
		rnd.updateRawString()
	case "uuid":
		panic("TODO: support uuid format")
	default:
		rnd.updateRawString()
	}
	return rnd.genString(format)
}

func (rnd Rnd) genInteger(format string) int64 {
	switch format {
	case "int32", "int64", "unixtime":
		return rnd.rawInteger
	default:
		return rnd.rawInteger
	}
}

func (rnd *Rnd) NextInteger(format string) int64 {
	rnd.updateRawInteger()
	return rnd.genInteger(format)
}

func (rnd Rnd) genNumber(format string) float64 {
	switch format {
	case "decimal", "double", "float":
		return rnd.rawNumber
	default:
		return rnd.rawNumber
	}
}

func (rnd *Rnd) NextNumber(format string) float64 {
	rnd.updateRawNumber()
	return rnd.genNumber(format)
}

func (rnd *Rnd) updateRawString() string {
	rl := []rune(rnd.rawString)
	for i := len(rl) - 1; i >= 0; i-- {
		if b := byte(rnd.rawString[i]); b != 'z' {
			rl[i] = rune(b + 1)
			rnd.rawString = string(rl)
			return rnd.rawString
		}
		rl[i] = 'a'
	}
	rnd.rawString = "a" + string(rl)
	return rnd.rawString
}

func (rnd *Rnd) updateRawInteger() int64 {
	rnd.rawInteger = rnd.rawInteger + 1
	return rnd.rawInteger
}

func (rnd *Rnd) updateRawNumber() float64 {
	rnd.rawNumber = rnd.rawNumber + 1
	return rnd.rawNumber
}

func (rnd *Rnd) nextRawTime(dur time.Duration) time.Time {
	rnd.time = rnd.time.Add(dur)
	return rnd.time
}
