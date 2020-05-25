package internal

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"io/ioutil"
	"sort"
)

// A Resource relates a bunch of bytes with a name and optionally other named resource forks, e.g. compressed
// variants of the same data.
type Resource struct {
	name  string
	blob  []byte
	forks []Resource
	size int64
}

// Name returns the resources unique name
func (r Resource) Name() string {
	return r.name
}

// Read opens the resource to read from. This never fails.
func (r Resource) Read() io.Reader {
	if r.blob == nil {
		gz := r.Fork("gzip")
		if gz != nil {
			r.blob = mustUnzip(gz.blob)
		} else {
			panic("internal error: embedded data provides neither default data nor a gzip fork")
		}
	}
	return bytes.NewReader(r.blob)
}

// Fork returns a named resource or nil if not found. Lookup is log(n).
func (r Resource) Fork(name string) *Resource {
	idx := sort.Search(len(r.forks), func(i int) bool {
		return r.forks[i].name == name
	})
	if idx < len(r.forks) && r.forks[idx].name == name {
		return &r.forks[idx]
	}
	return nil
}

// Resources contains all embedded named resources.
var resources []Resource = []Resource{
	{
		name: "",
		blob: mustUnzip(mustDecodeBase64("abcd")),
		forks: []Resource{
			{
				name: "bzip",
				blob: mustDecodeBase64("abc"),
			},
		},
	},
}

// mustDecodeBase64 panics, if str cannot be decoded
func mustDecodeBase64(str string) []byte {
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}

// mustUnzip panics, if b cannot be gunzipped
func mustUnzip(b []byte) []byte {
	reader, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	res, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return res
}


const sha2653324 = "asdv"