// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubeconfig

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestKubeconfig_ContextNames_noContextsEntry(t *testing.T) {
	tl := WithMockKubeconfigLoader(`a: b`)
	kc := New(tl)
	if err := kc.Parse(); err != nil {
		t.Fatal(err)
	}
	ctx := kc.ContextNames()
	var expected []string = nil
	if diff := cmp.Diff(expected, ctx); diff != "" {
		t.Fatalf("%s", diff)
	}
}

func TestKubeconfig_ContextNames_nonArrayContextsEntry(t *testing.T) {
	tl := WithMockKubeconfigLoader(`contexts: "hello"`)
	kc := New(tl)
	if err := kc.Parse(); err != nil {
		t.Fatal(err)
	}
	ctx := kc.ContextNames()
	var expected []string = nil
	if diff := cmp.Diff(expected, ctx); diff != "" {
		t.Fatalf("%s", diff)
	}
}
