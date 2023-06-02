package swagger

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
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
	uuid string
}

type RndOption struct {
	InitString  string
	InitInteger int64
	InitNumber  float64

	InitTime time.Time
	InitUUID string
}

func NewRnd(opt *RndOption) Rnd {
	if opt == nil {
		opt = &RndOption{
			InitString:  "a",
			InitInteger: 0,
			InitNumber:  0.5,
			InitTime:    time.Now(),
			InitUUID:    mustUUID(),
		}
	}
	return Rnd{
		rawString:  opt.InitString,
		rawInteger: opt.InitInteger,
		rawNumber:  opt.InitNumber,
		time:       opt.InitTime,
		uuid:       opt.InitUUID,
	}
}

func (rnd Rnd) RawString() string {
	return rnd.rawString
}

func (rnd *Rnd) NextRawString() string {
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

func (rnd Rnd) RawInteger() int64 {
	return rnd.rawInteger
}

func (rnd *Rnd) NextRawInteger() int64 {
	rnd.rawInteger = rnd.rawInteger + 1
	return rnd.rawInteger
}

func (rnd Rnd) RawNumber() float64 {
	return rnd.rawNumber
}

func (rnd *Rnd) NextRawNumber() float64 {
	rnd.rawNumber = rnd.rawNumber + 1
	return rnd.rawNumber
}

// Format: armid
func (rnd Rnd) ARMId() string {
	return "/subscriptions/00000000-0000-0000-000000000000/resourceGroups/" + rnd.rawString
}

func (rnd *Rnd) NextARMId() string {
	rnd.NextRawString()
	return rnd.ARMId()
}

// Format: base64url
func (rnd Rnd) Base64URL() string {
	return base64.StdEncoding.EncodeToString([]byte(rnd.rawString))
}

func (rnd *Rnd) NextBase64URL() string {
	rnd.NextRawString()
	return rnd.Base64URL()
}

// Format: binary
func (rnd Rnd) Binary() string {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, rnd.rawInteger)
	if err != nil {
		panic(fmt.Sprintf("binary.Write failed: %v", err))
	}
	return buf.String()
}

func (rnd *Rnd) NextBinary() string {
	rnd.NextRawInteger()
	return rnd.Binary()
}

// Format: byte
func (rnd Rnd) Byte() string {
	return rnd.Base64URL()
}

func (rnd *Rnd) NextByte() string {
	return rnd.NextBase64URL()
}

// Format: date
func (rnd Rnd) Date() string {
	return rnd.time.Format("2006-01-02")
}

func (rnd *Rnd) NextDate() string {
	rnd.time = rnd.time.Add(24 * time.Hour)
	return rnd.Date()
}

// Format: date-time
func (rnd Rnd) DateTime() string {
	return rnd.time.Format(time.RFC3339)
}

func (rnd *Rnd) NextDateTime() string {
	rnd.time = rnd.time.Add(time.Minute)
	return rnd.DateTime()
}

// Format: date-time-rfc1123
func (rnd Rnd) DateTimeRFC1123() string {
	return rnd.time.Format(time.RFC1123)
}

func (rnd *Rnd) NextDateTimeRFC1123() string {
	rnd.time = rnd.time.Add(time.Minute)
	return rnd.DateTimeRFC1123()
}

// Format: decimal
// Technically, we should use math/big, but the Azure SDK appears to be using the flaot64
func (rnd Rnd) Decimal() float64 {
	return rnd.RawNumber()
}

func (rnd *Rnd) NextDecimal() float64 {
	return rnd.NextRawNumber()
}

// Format: double
func (rnd Rnd) Double() float64 {
	return rnd.RawNumber()
}

func (rnd *Rnd) NextDouble() float64 {
	return rnd.NextRawNumber()
}

// Format: duration
func (rnd Rnd) Duration() string {
	p, _ := period.NewOf(time.Duration(rnd.rawInteger) * time.Second)
	return p.String()
}

func (rnd *Rnd) NextDuration() string {
	rnd.NextRawInteger()
	return rnd.Duration()
}

// Format: email
func (rnd Rnd) Email() string {
	return rnd.rawString + "@foo.com"
}

func (rnd *Rnd) NextEmail() string {
	rnd.NextRawString()
	return rnd.Email()
}

// Format: file
func (rnd Rnd) File() string {
	return rnd.RawString()
}

func (rnd Rnd) NextFile() string {
	return rnd.NextRawString()
}

// Format: float
func (rnd Rnd) Float() float32 {
	return float32(rnd.RawNumber())
}

func (rnd *Rnd) NextFloat() float32 {
	return float32(rnd.NextRawNumber())
}

// Format: int32
func (rnd Rnd) Int32() int32 {
	return int32(rnd.RawInteger())
}

func (rnd *Rnd) NextInt32() int32 {
	return int32(rnd.NextRawInteger())
}

// Format: int64
func (rnd Rnd) Int64() int64 {
	return rnd.RawInteger()
}

func (rnd *Rnd) NextInt64() int64 {
	return rnd.NextRawInteger()
}

// Format: password
func (rnd Rnd) Password() string {
	return rnd.RawString()
}

func (rnd *Rnd) NextPassword() string {
	return rnd.NextRawString()
}

// Format: time
func (rnd Rnd) Time() string {
	return rnd.time.Format("15:04:05")
}

func (rnd *Rnd) NextTime() string {
	rnd.time = rnd.time.Add(time.Minute)
	return rnd.Time()
}

// Format: unixtime
func (rnd Rnd) Unixtime() int64 {
	return rnd.RawInteger()
}

func (rnd *Rnd) NextUnixtime() int64 {
	return rnd.NextRawInteger()
}

// Format: uri
func (rnd Rnd) URI() string {
	return "https://" + rnd.RawString() + ".com"
}

func (rnd *Rnd) NextURI() string {
	rnd.NextRawString()
	return rnd.URI()
}

// Format: url
func (rnd Rnd) URL() string {
	return rnd.URI()
}

func (rnd *Rnd) NextURL() string {
	return rnd.NextURI()
}

// Format: uuid
func (rnd Rnd) UUID() string {
	return rnd.uuid
}

func (rnd *Rnd) NextUUID() string {
	rnd.uuid = mustUUID()
	return rnd.UUID()
}

func mustUUID() string {
	id, err := uuid.NewV4()
	if err != nil {
		panic(fmt.Sprintf("generating uuid: %v", err))
	}
	return id.String()
}
