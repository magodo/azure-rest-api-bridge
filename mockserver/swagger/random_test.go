package swagger

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRnd(t *testing.T) {
	initRnd := NewRnd(nil)
	cases := []struct {
		name   string
		expect interface{}
	}{
		{
			name:   "RawString",
			expect: "a",
		},
		{
			name:   "NextRawString",
			expect: "b",
		},
		{
			name:   "RawInteger",
			expect: int64(0),
		},
		{
			name:   "NextRawInteger",
			expect: int64(1),
		},
		{
			name:   "RawNumber",
			expect: 0.5,
		},
		{
			name:   "NextRawNumber",
			expect: 1.5,
		},
		{
			name:   "Date",
			expect: initRnd.time.Format("2006-01-02"),
		},
		{
			name:   "NextDate",
			expect: initRnd.time.Add(24 * time.Hour).Format("2006-01-02"),
		},
		{
			name:   "DateTime",
			expect: initRnd.time.Format(time.RFC3339),
		},
		{
			name:   "NextDateTime",
			expect: initRnd.time.Add(time.Minute).Format(time.RFC3339),
		},
		{
			name:   "DateTimeRFC1123",
			expect: initRnd.time.Format(time.RFC1123),
		},
		{
			name:   "NextDateTimeRFC1123",
			expect: initRnd.time.Add(time.Minute).Format(time.RFC1123),
		},
		{
			name:   "Decimal",
			expect: 0.5,
		},
		{
			name:   "NextDecimal",
			expect: 1.5,
		},
		{
			name:   "Double",
			expect: 0.5,
		},
		{
			name:   "NextDouble",
			expect: 1.5,
		},
		{
			name:   "Duration",
			expect: "P0D",
		},
		{
			name:   "NextDuration",
			expect: "PT1S",
		},
		{
			name:   "Email",
			expect: "a@foo.com",
		},
		{
			name:   "NextEmail",
			expect: "b@foo.com",
		},
		{
			name:   "File",
			expect: "a",
		},
		{
			name:   "NextFile",
			expect: "b",
		},
		{
			name:   "Float",
			expect: float32(0.5),
		},
		{
			name:   "NextFloat",
			expect: float32(1.5),
		},
		{
			name:   "Int32",
			expect: int32(0),
		},
		{
			name:   "NextInt32",
			expect: int32(1),
		},
		{
			name:   "Int64",
			expect: int64(0),
		},
		{
			name:   "NextInt64",
			expect: int64(1),
		},
		{
			name:   "Password",
			expect: "a",
		},
		{
			name:   "NextPassword",
			expect: "b",
		},
		{
			name:   "Time",
			expect: initRnd.time.Format("15:04:05"),
		},
		{
			name:   "NextTime",
			expect: initRnd.time.Add(time.Minute).Format("15:04:05"),
		},
		{
			name:   "Unixtime",
			expect: int64(0),
		},
		{
			name:   "NextUnixtime",
			expect: int64(1),
		},
		{
			name:   "URI",
			expect: "https://a.com",
		},
		{
			name:   "NextURI",
			expect: "https://b.com",
		},
		{
			name:   "URL",
			expect: "https://a.com",
		},
		{
			name:   "NextURL",
			expect: "https://b.com",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rnd := initRnd
			v := reflect.ValueOf(&rnd)
			mv := v.MethodByName(tt.name)
			rets := mv.Call(nil)
			require.Len(t, rets, 1)
			require.Equal(t, tt.expect, rets[0].Interface())
		})
	}
}

func TestRnd_NextRawStringCarray(t *testing.T) {
	rnd := NewRnd(nil)
	for i := 0; i != 26; i++ {
		rnd.NextRawString()
	}
	require.Equal(t, "aa", rnd.RawString())
}
