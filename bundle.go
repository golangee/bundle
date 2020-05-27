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
	"net/http"
	"sort"
)

// Bundle contains a bunch of resources.
type Bundle struct {
	resources []*Resource
}

// Make creates a new bundle from the given resources. We use Make here to avoid
// stuttering like bundle.NewBundle(). Takes ownership of resources.
func Make(resources ...*Resource) *Bundle {
	b := &Bundle{
		resources: resources,
	}
	b.sort()
	return b
}

// Handler returns a new http handler, providing resources for the given prefix
func (b *Bundle) Handler(prefix string) func(http.ResponseWriter, *http.Request) {
	return Handle(prefix, b.resources...)
}

// Put returns a new bundle instance with the given resource replacing any other resource
// with the same name. The current bundle is unchanged.
func (b *Bundle) Put(resource *Resource) *Bundle {
	idx, res := b.find(resource.name)
	if res != nil && idx < len(b.resources) {
		tmp := make([]*Resource, len(b.resources))
		copy(tmp, b.resources)
		tmp[idx] = resource
		return Make(tmp...)
	}

	tmp := append(b.resources[:idx], append([]*Resource{resource}, b.resources[idx:]...)...)
	return Make(tmp...)
}

// Remove tries to delete the resource from the bundle and returns a new potentionally modified instance.
func (b *Bundle) Remove(name string) *Bundle {
	idx, _ := b.find(name)
	if idx >= 0 {
		tmp := b.resources[:idx+copy(b.resources[idx:], b.resources[idx+1:])]
		return Make(tmp...)
	}
	return b
}

// Find returns the resource or nil
func (b *Bundle) Find(name string) *Resource {
	_, res := b.find(name)
	return res
}

// find returns the index and resource or nil
func (b *Bundle) find(name string) (int, *Resource) {
	idx := sort.Search(len(b.resources), func(i int) bool {
		return b.resources[i].name >= name
	})
	if idx < len(b.resources) && b.resources[idx].name == name {
		return idx, b.resources[idx]
	}
	return idx, nil
}

func (b *Bundle) sort() {
	sort.Sort(sortByName(b.resources))
}

type sortByName []*Resource

func (s sortByName) Len() int {
	return len(s)
}

func (s sortByName) Less(i, j int) bool {
	return s[i].name < s[j].name
}

func (s sortByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
