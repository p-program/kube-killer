now 		  := $(shell date)
PREFIX		  ?= zeusro
APP_NAME      ?= kube-killer:latest
IMAGE		  ?= $(PREFIX)/$(APP_NAME)
MIRROR_IMAGE  ?= registry.cn-shenzhen.aliyuncs.com/
ARCH		  ?=amd64

auto_commit:   
	git add .
	git commit -am "$(now)"
	git pull
	git push

build:
	GOARCH=$(ARCH) CGO_ENABLED=0 go build

build-kubectl-plugin:
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -o kubectl-kill ./cmd/kubectl-kill

buildAndRun: build run

fix-dep:
	go mod tidy
	go mod vendor

# usage dep="github.com/stretchr/testify/assert" make get
get:
	go get -u -v $(dep)
	go mod tidy
	go mod vendor

mirror: pull
	docker build -t $(MIRROR_IMAGE) -f deploy/docker/Dockerfile .

pull:
	git reset --hard HEAD
	git pull

release-mirror: mirror
	docker push $(MIRROR_IMAGE)

rebuild: pull	
	docker build -t $(IMAGE) -f deploy/docker/Dockerfile .

run:
	./kube-killer

test:
	mkdir -p artifacts/report/coverage
	go test -v -cover -coverprofile c.out.tmp ./...
	cat c.out.tmp | grep -v "_mock.go" > c.out
	go tool cover -html=c.out -o artifacts/report/coverage/index.html	

up:
	docker-compose build --force-rm --no-cache
	docker-compose up

update-dep: update-mod fix-dep

update-mod:
	go get -u -v github.com/p-program/go-common-library
	# go-mod-upgrade

# Install kubectl-kill plugin to ~/bin or /usr/local/bin
# Usage: make install-kubectl-plugin [PREFIX=/usr/local]
install-kubectl-plugin: build-kubectl-plugin
	@if [ -z "$(PREFIX)" ]; then \
		PREFIX=$$HOME/bin; \
	fi; \
	mkdir -p $$PREFIX; \
	cp kubectl-kill $$PREFIX/; \
	chmod +x $$PREFIX/kubectl-kill; \
	echo "kubectl-kill plugin installed to $$PREFIX/kubectl-kill"; \
	echo "Make sure $$PREFIX is in your PATH"
