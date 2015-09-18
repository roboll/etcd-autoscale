###############################################################################
# make tasks for building and uploading release artifacts to github releases
###############################################################################
all: $(PRE_RELEASE)

OWNER   := roboll
REPO    := etcd-autoscale-members
PROJECT := github.com/$(OWNER)/$(REPO)
VERSION := $(shell git describe --tags)

###############################################################################
# pre-release - test and validation steps
###############################################################################
PRE_RELEASE := test

.PHONY: test
test: ; go test ./...

###############################################################################
# build-release - build a release artifact
###############################################################################
GOOS     := linux
GOARCH   := amd64
ARTIFACT := $(REPO)-$(GOOS)-$(GOARCH)-$(VERSION)

.PHONY: build build-release
build-release:
	docker run \
		-v $(PWD):/go/src/$(PROJECT) -v $(PWD)/target:/target \
		golang /bin/bash -c \
			"CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
			go get $(PROJECT)/... && \
			go build -a -tags netgo -ldflags '-s -w' \
			-o /target/$(ARTIFACT) $(PROJECT)"

###############################################################################
# perform-release - upload a release to github releases
#
# requirements:
# - the checked out revision be a pushed tag
# - a github api token (GITHUB_TOKEN)
###############################################################################
API    = https://api.github.com/repos/$(OWNER)/$(REPO)
UPLOAD = https://uploads.github.com/repos/$(OWNER)/$(REPO)

.PHONY: tag github-token create-release perform-release
tag: ; git describe --tags --exact-match HEAD

github-token:
ifndef GITHUB_TOKEN
	$(error $GITHUB_TOKEN not set)
endif

create-release: tag github-token
	curl -s -XPOST -H "Authorization: token $(GITHUB_TOKEN)" \
		$(API)/releases -d '{ "tag_name": "$(VERSION)" }'

perform-release: tag github-token $(PRE_RELEASE) build-release create-release
	curl -s -H "Authorization: token $(GITHUB_TOKEN)" \
		$(API)/releases/tags/$(VERSION) | \
	python -c "import json,sys;obj=json.load(sys.stdin);print obj['id']" | \
		curl -s -XPOST \
			-H "Authorization: token $(GITHUB_TOKEN)" \
			-H "Content-Type: application/octet-stream" \
			$(UPLOAD)/releases/$$(xargs )/assets?name=$(ARTIFACT) \
			--data-binary @target/$(ARTIFACT)
