all: check format lint build

build: render-ci-pipeline

render-ci-pipeline:
    ./scripts/render-ci-pipeline.sh

fmt: format
format: format-dhall prettier format-shfmt format-golang

lint: lint-dhall shellcheck

check: check-dhall test

prettier:
    yarn run prettier

format-dhall:
    ./scripts/dhall-format.sh

check-dhall:
    ./scripts/dhall-check.sh

lint-dhall:
    ./scripts/dhall-lint.sh

shellcheck:
    ./scripts/shellcheck.sh

format-shfmt:
    shfmt -w .

gofmt: format-golang
format-golang:
    ./scripts/go-fmt.sh

test: test-golang

test-golang:
    ./scripts/go-test.sh

install:
    just install-asdf
    just install-yarn

install-yarn:
    yarn

install-asdf:
    asdf install
