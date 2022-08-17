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
	"encoding/json"
	"sort"
	"strings"
)

type StringSet struct {
	M     map[string]struct{}
	owned bool
}

func NewStringSet(s ...string) StringSet {
	ret := StringSet{M: make(map[string]struct{}), owned: true}
	for _, x := range s {
		ret.M[x] = struct{}{}
	}
	return ret
}

func (set StringSet) Clone() StringSet {
	return StringSet{M: set.M}
}

func (set StringSet) IsEqual(s StringSet) bool {
	return len(set.M) == len(s.M) && set.HasAllSet(s)
}

func (set StringSet) Has(s string) bool {
	_, ok := set.M[s]
	return ok
}

func (set StringSet) HasAny(s ...string) bool {
	for _, x := range s {
		if _, ok := set.M[x]; ok {
			return true
		}
	}
	return false
}

func (set StringSet) HasAnySet(s StringSet) bool {
	for x := range s.M {
		if _, ok := set.M[x]; ok {
			return true
		}
	}
	return false
}

func (set StringSet) HasAll(s ...string) bool {
	for _, x := range s {
		if _, ok := set.M[x]; !ok {
			return false
		}
	}
	return true
}

func (set StringSet) HasAllSet(s StringSet) bool {
	for x := range s.M {
		if _, ok := set.M[x]; !ok {
			return false
		}
	}
	return true
}

func (set *StringSet) Add(s ...string) *StringSet {
	set.own()
	for _, x := range s {
		set.M[x] = struct{}{}
	}
	return set
}

func (set *StringSet) AddSet(s StringSet) *StringSet {
	set.own()
	for x := range s.M {
		set.M[x] = struct{}{}
	}
	return set
}

func (set *StringSet) Remove(s ...string) *StringSet {
	set.own()
	for _, x := range s {
		delete(set.M, x)
	}
	return set
}

func (set StringSet) Slice() []string {
	ret := make([]string, 0, len(set.M))
	for k := range set.M {
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

func (set StringSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(set.Slice())
}

func (set *StringSet) UnmarshalJSON(in []byte) error {
	var arr []string
	if err := json.Unmarshal(in, &arr); err != nil {
		return err
	}
	*set = NewStringSet(arr...)
	return nil
}

func (set StringSet) Len() int { return len(set.M) }

func (set *StringSet) own() {
	if set.owned {
		return
	}
	m := make(map[string]struct{}, len(set.M))
	for x := range set.M {
		m[x] = struct{}{}
	}
	set.M = m
	set.owned = true
}
