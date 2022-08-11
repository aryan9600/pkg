/*
Copyright 2022 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package git

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCleanPath(t *testing.T) {
	g := NewWithT(t)

	tmp := t.TempDir()
	cleaned, err := CleanPath(tmp)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cleaned).To(Equal(tmp))

	wd, err := os.Getwd()
	g.Expect(err).ToNot(HaveOccurred())

	rel := "./relative"
	cleaned, err = CleanPath(rel)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cleaned).To(Equal(filepath.Join(wd, "relative")))

	base := "../../outside"
	cleaned, err = CleanPath(base)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cleaned).To(Equal(filepath.Join(wd, "outside")))
}
