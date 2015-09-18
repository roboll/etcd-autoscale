###############################################################################
# release tasks for binary and docker releases
###############################################################################
OWNER     := roboll
REPO      := etcd-autoscale-members

PROJECT   := github.com/$(OWNER)/$(REPO)
IMAGE_TAG := $(OWNER)/$(REPO):$(VERSION)
VERSION   := $(shell git describe --tags)

all: $(PRE_RELEASE)
release: $(PRE_RELEASE) perform-binary-release perform-docker-release

###############################################################################
# pre-release - test and validation steps
###############################################################################
PRE_RELEASE := test

.PHONY: test
test: ; go test ./...

###############################################################################
# binary-release - build a binary release artifact
###############################################################################
GOOS     := linux
GOARCH   := amd64
ARTIFACT := $(REPO)-$(GOOS)-$(GOARCH)-$(VERSION)

.PHONY: build-binary-release
build-binary-release:
	docker run \
		-v $(PWD):/go/src/$(PROJECT) -v $(PWD)/target:/target \
		golang /bin/bash -c \
			"CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
			go get $(PROJECT)/... && \
			go build -a -tags netgo -ldflags '-s -w' \
			-o /target/$(ARTIFACT) $(PROJECT)"

###############################################################################
# github-release - upload a binary release to github releases
#
# requirements:
# - the checked out revision be a pushed tag
# - a github api token (GITHUB_TOKEN)
###############################################################################
API    = https://api.github.com/repos/$(OWNER)/$(REPO)
UPLOAD = https://uploads.github.com/repos/$(OWNER)/$(REPO)

.PHONY: create-gh-release perform-gh-release gh-token
create-gh-release: tag clean gh-token
	$(info Creating Github Release)
	@curl -s -XPOST -H "Authorization: token $(GITHUB_TOKEN)" \
		$(API)/releases -d '{ "tag_name": "$(VERSION)" }' > /dev/null

perform-gh-release: tag clean gh-token build-binary-release create-gh-release
	$(info Uploading Release Artifact to Github)
	@curl -s \
		-H "Authorization: token $(GITHUB_TOKEN)" \
		$(API)/releases/tags/$(VERSION) |\
	python -c "import json,sys;obj=json.load(sys.stdin);print obj['id']" |\
	curl -s -XPOST \
		-H "Authorization: token $(GITHUB_TOKEN)" \
		-H "Content-Type: application/octet-stream" \
		$(UPLOAD)/releases/$$(xargs )/assets?name=$(ARTIFACT) \
		--data-binary @target/$(ARTIFACT) > /dev/null

gh-token:
ifndef GITHUB_TOKEN
	$(error $GITHUB_TOKEN not set)
endif

###############################################################################
# docker-release - push a docker image release
###############################################################################
build-docker-release: Dockerfile build-binary-release
	docker build -t $(IMAGE_TAG) .

perform-docker-release: Dockerfile tag clean build-docker-release
	docker push $(IMAGE_TAG)

###############################################################################
# utility
###############################################################################
.PHONY: tag clean
tag:   ; @git describe --tags --exact-match HEAD
clean:
	@git diff --exit-code > /dev/null
	@git diff --cached --exit-code > /dev/null
