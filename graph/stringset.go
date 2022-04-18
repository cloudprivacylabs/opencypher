// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package graph

import (
	"sort"
	"strings"
)

type StringSet map[string]struct{}

func NewStringSet(s ...string) StringSet {
	ret := make(StringSet)
	for _, x := range s {
		ret[x] = struct{}{}
	}
	return ret
}

func (set StringSet) Clone() StringSet {
	ret := make(StringSet)
	for x := range set {
		ret[x] = struct{}{}
	}
	return ret
}

func (set StringSet) IsEqual(s StringSet) bool {
	return len(set) == len(s) && set.HasAllSet(s)
}

func (set StringSet) Has(s string) bool {
	_, ok := set[s]
	return ok
}

func (set StringSet) HasAny(s ...string) bool {
	for _, x := range s {
		if _, ok := set[x]; ok {
			return true
		}
	}
	return false
}

func (set StringSet) HasAnySet(s StringSet) bool {
	for x := range s {
		if _, ok := set[x]; ok {
			return true
		}
	}
	return false
}

func (set StringSet) HasAll(s ...string) bool {
	for _, x := range s {
		if _, ok := set[x]; !ok {
			return false
		}
	}
	return true
}

func (set StringSet) HasAllSet(s StringSet) bool {
	for x := range s {
		if _, ok := set[x]; !ok {
			return false
		}
	}
	return true
}

func (set StringSet) Add(s ...string) StringSet {
	for _, x := range s {
		set[x] = struct{}{}
	}
	return set
}

func (set StringSet) AddSet(s StringSet) StringSet {
	for x := range s {
		set[x] = struct{}{}
	}
	return set
}

func (set StringSet) Remove(s ...string) StringSet {
	for _, x := range s {
		delete(set, x)
	}
	return set
}

func (set StringSet) Slice() []string {
	ret := make([]string, 0, len(set))
	for k := range set {
		ret = append(ret, k)
	}
	return ret
}

func (set StringSet) SortedSlice() []string {
	ret := set.Slice()
	sort.Strings(ret)
	return ret
}

func (set StringSet) String() string {
	return strings.Join(set.Slice(), ", ")
}
