package refutil

import (
	"testing"

	"github.com/go-openapi/jsonpointer"
	"github.com/stretchr/testify/assert"
)

func TestJSONPointerOffsetMulti(t *testing.T) {
	cases := []struct {
		name     string
		ptr      []string
		input    string
		offset   map[string]int64
		hasError bool
	}{
		{
			name: "object key",
			ptr: []string{
				"/foo/bar",
				"/foo/baz",
				"/foo",
			},
			input: `{"foo": {"bar": 21, "baz": 42, "bed": 11}, "zoo":"1"}`,
			offset: map[string]int64{
				"/foo/bar": 9,
				"/foo/baz": 18,
				"/foo":     1,
			},
		},
		{
			name: "array index",
			ptr: []string{
				"/0/0",
				"/0/1",
				"/1/0",
				"/1/1",
			},
			input: `[[1,2], [3,4]]`,
			offset: map[string]int64{
				"/0/0": 2,
				"/0/1": 3,
				"/1/0": 9,
				"/1/1": 10,
			},
		},
		{
			name: "mix array index and object key",
			ptr: []string{
				"/0/1/foo/0",
			},
			input: `[[1, {"foo": ["a", "b"]}], [3, 4]]`,
			offset: map[string]int64{
				"/0/1/foo/0": 14,
			},
		},
		{
			name: "nonexist object key",
			ptr: []string{
				"/foo/baz",
			},
			input:    `{"foo": {"bar": 21}}`,
			hasError: true,
		},
		{
			name: "nonexist array index",
			ptr: []string{
				"/0/2",
			},
			input:    `[[1,2], [3,4]]`,
			hasError: true,
		},
		{
			name: "encoded reference",
			ptr: []string{
				"/paths/~1p~1{}/get",
			},
			input: `{"paths": {"foo": {"bar": 123, "baz": {}}, "/p/{}": {"get": {}}}}`,
			offset: map[string]int64{
				"/paths/~1p~1{}/get": 53,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ptrs := make([]jsonpointer.Pointer, 0)
			for _, p := range tt.ptr {
				ptr, err := jsonpointer.New(p)
				assert.NoError(t, err)
				ptrs = append(ptrs, ptr)
			}
			offset, err := JSONPointerOffsetMulti(ptrs, tt.input)
			if tt.hasError {
				assert.Error(t, err)
				return
			}
			t.Log(offset, err)
			assert.NoError(t, err)
			assert.Equal(t, tt.offset, offset)
		})
	}
}
