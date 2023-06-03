package swagger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRnd(t *testing.T) {
	initRnd := NewRnd(nil)
	cases := []struct {
		typ    string
		format string
		again  bool
		expect interface{}
	}{
		{
			typ:    "string",
			format: "",
			expect: "b",
		},
		{
			typ:    "string",
			format: "",
			again:  true,
			expect: "c",
		},
		{
			typ:    "integer",
			format: "",
			expect: int64(1),
		},
		{
			typ:    "integer",
			format: "",
			again:  true,
			expect: int64(2),
		},
		{
			typ:    "number",
			format: "",
			expect: 1.5,
		},
		{
			typ:    "number",
			format: "",
			again:  true,
			expect: 2.5,
		},
		{
			typ:    "string",
			format: "date",
			expect: initRnd.time.Add(24 * time.Hour).Format("2006-01-02"),
		},
		{
			typ:    "string",
			format: "date",
			again:  true,
			expect: initRnd.time.Add(48 * time.Hour).Format("2006-01-02"),
		},
		{
			typ:    "string",
			format: "date-time",
			expect: initRnd.time.Add(time.Hour).Format(time.RFC3339),
		},
		{
			typ:    "string",
			format: "date-time",
			again:  true,
			expect: initRnd.time.Add(time.Duration(2) * time.Hour).Format(time.RFC3339),
		},
		{
			typ:    "string",
			format: "date-time-rfc1123",
			expect: initRnd.time.Add(time.Hour).Format(time.RFC1123),
		},
		{
			typ:    "string",
			format: "date-time-rfc1123",
			again:  true,
			expect: initRnd.time.Add(time.Duration(2) * time.Hour).Format(time.RFC1123),
		},
		{
			typ:    "number",
			format: "decimal",
			expect: 1.5,
		},
		{
			typ:    "number",
			format: "decimal",
			again:  true,
			expect: 2.5,
		},
		{
			typ:    "number",
			format: "double",
			expect: 1.5,
		},
		{
			typ:    "number",
			format: "double",
			again:  true,
			expect: 2.5,
		},
		{
			typ:    "string",
			format: "duration",
			expect: "PT1H",
		},
		{
			typ:    "string",
			format: "duration",
			again:  true,
			expect: "PT2H",
		},
		{
			typ:    "string",
			format: "email",
			expect: "b@foo.com",
		},
		{
			typ:    "string",
			format: "email",
			again:  true,
			expect: "c@foo.com",
		},
		{
			typ:    "string",
			format: "file",
			expect: "b",
		},
		{
			typ:    "string",
			format: "file",
			again:  true,
			expect: "c",
		},
		{
			typ:    "number",
			format: "float",
			expect: 1.5,
		},
		{
			typ:    "number",
			format: "float",
			again:  true,
			expect: 2.5,
		},
		{
			typ:    "integer",
			format: "int32",
			expect: int64(1),
		},
		{
			typ:    "integer",
			format: "int32",
			again:  true,
			expect: int64(2),
		},
		{
			typ:    "integer",
			format: "int64",
			expect: int64(1),
		},
		{
			typ:    "integer",
			format: "int64",
			again:  true,
			expect: int64(2),
		},
		{
			typ:    "string",
			format: "password",
			expect: "b",
		},
		{
			typ:    "string",
			format: "password",
			again:  true,
			expect: "c",
		},
		{
			typ:    "string",
			format: "time",
			expect: initRnd.time.Add(time.Hour).Format("15:04:05"),
		},
		{
			typ:    "string",
			format: "time",
			again:  true,
			expect: initRnd.time.Add(time.Duration(2) * time.Hour).Format("15:04:05"),
		},
		{
			typ:    "integer",
			format: "unixtime",
			expect: int64(1),
		},
		{
			typ:    "integer",
			format: "unixtime",
			again:  true,
			expect: int64(2),
		},
		{
			typ:    "string",
			format: "uri",
			expect: "https://b.com",
		},
		{
			typ:    "string",
			format: "uri",
			again:  true,
			expect: "https://c.com",
		},
		{
			typ:    "string",
			format: "url",
			expect: "https://b.com",
		},
		{
			typ:    "string",
			format: "url",
			again:  true,
			expect: "https://c.com",
		},
	}

	for _, tt := range cases {
		name := tt.typ
		if tt.format != "" {
			name += "(" + tt.format + ")"
		}
		if tt.again {
			name = "next " + name
		}
		t.Run(name, func(t *testing.T) {
			rnd := initRnd
			var v interface{}
			switch tt.typ {
			case "string":
				v = rnd.NextString(tt.format)
				if tt.again {
					v = rnd.NextString(tt.format)
				}
			case "integer":
				v = rnd.NextInteger(tt.format)
				if tt.again {
					v = rnd.NextInteger(tt.format)
				}
			case "number":
				v = rnd.NextNumber(tt.format)
				if tt.again {
					v = rnd.NextNumber(tt.format)
				}
			default:
				t.Fatalf("unknown type: %s", tt.typ)
			}
			require.Equal(t, tt.expect, v)
		})
	}
}

func TestRnd_NextRawStringCarray(t *testing.T) {
	rnd := NewRnd(nil)
	for i := 0; i != 26; i++ {
		rnd.updateRawString()
	}
	require.Equal(t, "aa", rnd.rawString)
}
