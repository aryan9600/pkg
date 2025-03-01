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

package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	gcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	. "github.com/onsi/gomega"
)

func Test_List(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	c := NewLocalClient()
	repo := "test-list" + randStringRunes(5)
	tags := []string{"v0.0.1", "v0.0.2", "v0.0.3", "v6.0.0", "v6.0.1", "v6.0.2", "v6.0.2-rc.1", "v6.0.2-alpha", "staging-fb3355b"}
	source := "github.com/fluxcd/fluxv2"
	rev := "rev"
	ct := time.Now()
	m := Metadata{
		Source:   source,
		Revision: rev,
		Created:  ct.Format(time.RFC3339),
	}

	for _, tag := range tags {
		dst := fmt.Sprintf("%s/%s:%s", dockerReg, repo, tag)
		img, err := random.Image(1024, 1)
		g.Expect(err).ToNot(HaveOccurred())
		img = mutate.Annotations(img, m.ToAnnotations()).(gcrv1.Image)
		err = crane.Push(img, dst, c.options...)
		g.Expect(err).ToNot(HaveOccurred())
	}

	tests := []struct {
		name         string
		regexFilter  string
		semverFilter string
		expectedTags []string
	}{
		{
			name:         "list all tags",
			expectedTags: tags,
		},
		{
			name:         "list 6.0.x tags",
			semverFilter: "6.0.x",
			expectedTags: []string{"v6.0.0", "v6.0.1", "v6.0.2"},
		},
		{
			name:         "list tags starting with staging",
			regexFilter:  "^staging.*$",
			expectedTags: []string{"staging-fb3355b"},
		},
		{
			name:         "filter for alpha 6.0.x",
			semverFilter: ">=6.0.0-rc",
			regexFilter:  "alpha",
			expectedTags: []string{"v6.0.2-alpha"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			metadata, err := c.List(ctx, fmt.Sprintf("%s/%s", dockerReg, repo), ListOptions{
				SemverFilter: tt.semverFilter,
				RegexFilter:  tt.regexFilter,
			})
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(len(metadata)).To(Equal(len(tt.expectedTags)))
			for _, meta := range metadata {
				tag, err := name.NewTag(meta.URL)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tag.TagStr()).Should(BeElementOf(tt.expectedTags))

				g.Expect(meta.ToAnnotations()).To(Equal(m.ToAnnotations()))

				digest, err := crane.Digest(meta.URL, c.options...)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(meta.Digest).To(Equal(digest))
			}
		})
	}
}
