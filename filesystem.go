// Copyright 2020 Torben Schinke
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bundle

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Filesystem is an implementation on top of the bundle resources. However better use the handler, which returns
// handles compression better.
type Filesystem struct {
	entries map[string]*fileResourceAdapter
}

func NewFilesystem(resources ...*Resource) *Filesystem {
	f := &Filesystem{entries: map[string]*fileResourceAdapter{}}
	for _, r := range resources {
		paths := strings.Split(r.name, "/")
		name := paths[len(paths)-1]
		parentName := ""
		if len(paths) > 1 {
			parentName = paths[len(paths)-2]
		}
		parent := "/" + strings.Join(paths[:len(paths)-1], "/")

		p := f.entries[parent]
		if p == nil {
			p = &fileResourceAdapter{name: parentName, children: []*Resource{}}
			f.entries[parent] = p
		}
		p.children = append(p.children, r)

		f.entries["/"+strings.Join(paths, "/")] = &fileResourceAdapter{
			name:     name,
			res:      r,
			children: nil,
			seeker:   nil,
		}
	}
	return f
}

func (b Filesystem) Open(name string) (http.File, error) {
	f := b.entries[name]
	if f == nil {
		return nil, fmt.Errorf("not found")
	}
	return f, nil
}

type fileResourceAdapter struct {
	name     string
	res      *Resource
	children []*Resource
	seeker   *byteSeeker
}

func (f *fileResourceAdapter) Close() error {
	return nil
}

func (f *fileResourceAdapter) getSeeker() *byteSeeker {
	if f.seeker == nil {
		f.seeker = &byteSeeker{buf: f.res.unpack()}
	}
	return f.seeker
}

func (f *fileResourceAdapter) Read(p []byte) (n int, err error) {
	return f.getSeeker().Read(p)
}

func (f *fileResourceAdapter) Seek(offset int64, whence int) (int64, error) {
	return f.getSeeker().Seek(offset, whence)
}

func (f *fileResourceAdapter) IsDir() bool {
	return f.children != nil
}

func (f *fileResourceAdapter) Readdir(count int) ([]os.FileInfo, error) {
	if !f.IsDir() {
		return nil, &os.PathError{
			Op:   "Readdir",
			Path: f.name,
			Err:  fmt.Errorf("is a file"),
		}
	}
	res := make([]os.FileInfo, len(f.children))
	for i, r := range f.children {
		res[i] = r
	}
	return res, nil
}

func (f *fileResourceAdapter) Stat() (os.FileInfo, error) {
	if f.IsDir() {
		return dummyDir{name: f.name}, nil
	}
	return f.res, nil
}

type dummyDir struct {
	name string
}

func (d dummyDir) Name() string {
	return d.name
}

func (d dummyDir) Size() int64 {
	return 0
}

func (d dummyDir) Mode() os.FileMode {
	return os.ModePerm
}

func (d dummyDir) ModTime() time.Time {
	return time.Unix(0, 0)
}

func (d dummyDir) IsDir() bool {
	return true
}

func (d dummyDir) Sys() interface{} {
	return nil
}
