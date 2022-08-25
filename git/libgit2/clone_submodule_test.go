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

package libgit2

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git2go "github.com/libgit2/git2go/v33"
	. "github.com/onsi/gomega"

	"github.com/fluxcd/pkg/git"
	"github.com/fluxcd/pkg/git/libgit2/internal/test"
	"github.com/fluxcd/pkg/git/libgit2/transport"
	"github.com/fluxcd/pkg/gittestserver"
)

func Test_updateSubmodules(t *testing.T) {
	server, err := gittestserver.NewTempGitServer()
	if err != nil {
		t.Fatal(err)
	}
	// defer os.RemoveAll(server.Root())

	err = server.StartHTTP()
	if err != nil {
		t.Fatal(err)
	}
	defer server.StopHTTP()
	fmt.Println(server.Root())

	baseRepoPath := "base.git"
	err = server.InitRepo("../testdata/git/repo", git.DefaultBranch, baseRepoPath)
	if err != nil {
		t.Fatal(err)
	}

	icingRepoPath := "icing.git"
	err = server.InitRepo("../testdata/git/repo2", git.DefaultBranch, icingRepoPath)
	if err != nil {
		t.Fatal(err)
	}

	toppingsRepoPath := "toppings.git"
	err = server.InitRepo("../testdata/git/repo3", git.DefaultBranch, toppingsRepoPath)
	if err != nil {
		t.Fatal(err)
	}

	toppingsRepoURL := server.HTTPAddress() + "/" + toppingsRepoPath
	icingRepoURL := server.HTTPAddress() + "/" + icingRepoPath
	baseRepoURL := server.HTTPAddress() + "/" + baseRepoPath

	icingRepo, err := clone(context.TODO(), icingRepoURL, git.DefaultBranch, &git.AuthOptions{
		Transport: git.HTTP,
	})
	defer icingRepo.Free()
	defer os.RemoveAll(icingRepo.Workdir())

	err = addSubmodule(icingRepo, baseRepoURL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = commit(icingRepo, "add base submodule")
	if err != nil {
		t.Fatal(err)
	}
	err = push(icingRepo, git.DefaultBranch, &git.AuthOptions{
		Transport: git.HTTP,
	})
	if err != nil {
		t.Fatal(err)
	}

	toppingsRepo, err := clone(context.TODO(), toppingsRepoURL, git.DefaultBranch, &git.AuthOptions{
		Transport: git.HTTP,
	})
	defer os.RemoveAll(toppingsRepo.Workdir())

	err = addSubmodule(toppingsRepo, icingRepoURL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = commit(toppingsRepo, "add icing submodule")
	if err != nil {
		t.Fatal(err)
	}
	err = push(toppingsRepo, git.DefaultBranch, &git.AuthOptions{
		Transport: git.HTTP,
	})
	if err != nil {
		t.Fatal(err)
	}

	g := NewWithT(t)
	g.Expect(err).ToNot(HaveOccurred())
	tmp := t.TempDir()
	lgc, err := NewClient(tmp, &git.AuthOptions{
		Transport: git.HTTP,
		// Username:  "git",
	})
	g.Expect(err).ToNot(HaveOccurred())
	defer lgc.Close()

	_, err = lgc.Clone(context.TODO(), icingRepoURL, git.CheckoutOptions{
		Branch:            git.DefaultBranch,
		RecurseSubmodules: true,
	})
	g.Expect(err).ToNot(HaveOccurred())

	// lgc.repository = toppingsRepo
	// remote, err := toppingsRepo.Remotes.Lookup(git.DefaultRemote)
	// g.Expect(err).ToNot(HaveOccurred())
	// lgc.remote = remote

	// err = lgc.updateSubmodules()
	// time.Sleep(time.Second * 60)
	// g.Expect(err).ToNot(HaveOccurred())

	filepath.WalkDir(lgc.Path(), func(path string, d fs.DirEntry, err error) error {
		if !strings.Contains(path, ".git") {
			fmt.Println(path)
		}
		return nil
	})
	// cmd := exec.Command("git", "clone", icingRepoURL)
	// var stdout bytes.Buffer
	// var stderr bytes.Buffer
	// cmd.Stdout = &stdout
	// cmd.Stderr = &stderr
	// if err := cmd.Run(); err != nil {
	// fmt.Println(stdout.String(), stderr.String())
	// }
}

func addSubmodule(repo *git2go.Repository, submoduleURL string) error {
	cmd := exec.Command("git", "submodule", "add", submoduleURL)
	cmd.Dir = repo.Workdir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	fmt.Printf("adding submodule: %s\n", submoduleURL)
	fmt.Println(stdout.String(), stderr.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error while adding submodule: %s: %w\n%s\n%s",
			submoduleURL, err, stdout.String(), stderr.String())
	}
	filepath.WalkDir(cmd.Dir, func(path string, d fs.DirEntry, err error) error {
		if !strings.Contains(path, ".git") {
			fmt.Println(path)
		}
		return nil
	})
	return nil
}

func clone(ctx context.Context, repoURL, branchName string, authOpts *git.AuthOptions) (*git2go.Repository, error) {
	dir, err := os.MkdirTemp("", "git-libgit2-clone-")
	if err != nil {
		return nil, err
	}

	transportOptsURL := getTransportOptsURL(authOpts.Transport)
	transport.AddTransportOptions(transportOptsURL, transport.TransportOptions{
		TargetURL: repoURL,
		Context:   ctx,
		AuthOpts:  authOpts,
	})
	defer transport.RemoveTransportOptions(transportOptsURL)

	opts := &git2go.CloneOptions{
		Bare:           false,
		CheckoutBranch: branchName,
		CheckoutOptions: git2go.CheckoutOptions{
			Strategy: git2go.CheckoutForce,
		},
	}
	repo, err := git2go.Clone(transportOptsURL, dir, opts)
	if err != nil {
		return nil, err
	}

	// set the origin remote url to the actual repo url, since
	// the origin remote will have transportOptsURl as the it's url
	// because that's the url used to clone the repo.
	err = repo.Remotes.SetUrl(git.DefaultRemote, repoURL)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func push(repo *git2go.Repository, branch string, authOpts *git.AuthOptions) error {
	origin, err := repo.Remotes.Lookup(git.DefaultRemote)
	if err != nil {
		return err
	}
	defer origin.Free()

	url := origin.Url()
	transportOptsURL := getTransportOptsURL(authOpts.Transport)
	transport.AddTransportOptions(transportOptsURL, transport.TransportOptions{
		TargetURL: url,
		Context:   context.TODO(),
		AuthOpts:  authOpts,
	})
	defer transport.RemoveTransportOptions(transportOptsURL)

	err = repo.Remotes.SetUrl(git.DefaultRemote, transportOptsURL)
	if err != nil {
		return err
	}

	origin, err = repo.Remotes.Lookup(git.DefaultRemote)
	if err != nil {
		return err
	}

	err = origin.Push([]string{fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)}, &git2go.PushOptions{
		RemoteCallbacks: RemoteCallbacks(),
		ProxyOptions:    git2go.ProxyOptions{Type: git2go.ProxyTypeAuto},
	})
	if err != nil {
		return err
	}

	repo.Remotes.SetUrl(git.DefaultRemote, url)
	return nil
}

func commit(repo *git2go.Repository, message string) (string, error) {
	var parentC []*git2go.Commit
	head, err := test.HeadCommit(repo)
	if err == nil {
		defer head.Free()
		parentC = append(parentC, head)
	}

	index, err := repo.Index()
	if err != nil {
		return "", err
	}

	// err = index.AddAll(nil, git2go.IndexAddDefault, func(s1, s2 string) error {
	// return nil
	// })
	// if err != nil {
	// fmt.Println(err)
	// return "", err
	// }

	if err := index.Write(); err != nil {
		return "", err
	}

	treeID, err := index.WriteTree()
	if err != nil {
		return "", err
	}

	tree, err := repo.LookupTree(treeID)
	if err != nil {
		return "", err
	}
	defer tree.Free()

	sig := &git2go.Signature{
		Name:  "Test User",
		Email: "test@user.com",
		When:  time.Now(),
	}

	commitID, err := repo.CreateCommit("HEAD", sig, sig, message, tree, parentC...)
	if err != nil {
		return "", err
	}

	return commitID.String(), nil
}
