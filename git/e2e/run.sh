#!/bin/bash
# This script runs e2e tests for pkg/git.

DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(git rev-parse --show-toplevel)"

pushd $PROJECT_ROOT/git && make libgit2
popd

GOLANG_WITH_LIBGIT2_VER="v0.2.0"
LIBGIT2_PC=$PROJECT_ROOT/git/libgit2/build/libgit2/$GOLANG_WITH_LIBGIT2_VER/lib/pkgconfig/
CGO_LDFLAGS=$(PKG_CONFIG_PATH=$LIBGIT2_PC pkg-config --libs --static --cflags libgit2 2>/dev/null)

source "$DIR"/setup_gitlab.sh
PKG_CONFIG_PATH=$LIBGIT2_PC CGO_LDFLAGS=$CGO_LDFLAGS go test -v -tags 'netgo,osusergo,static_build,e2e' -race ./...

# cleanup
docker kill gitlab && docker rm gitlab

