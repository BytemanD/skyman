
UPX=$(shell which upx 2> /dev/null)
VERSION := $(shell go run cmd/skyman.go -v |awk '{{print $$3}}')
GO_VERSION := $(shell go version |awk '{{print $$3}}')
BUILD_DATE := $(shell date +'%Y-%m-%d %H:%M:%S')
UNAME := $(shell uname -si)

BUILD_SPEC=/tmp/skyman.spec

build:
	go mod download
	mkdir -p dist
	go build  -o dist/ -ldflags " \
		-X 'main.Version=$(VERSION)' \
		-X 'main.GoVersion=$(GO_VERSION)' \
		-X 'main.BuildDate=$(BUILD_DATE)' \
		-X 'main.BuildPlatform=$(UNAME)' -s -w" \
		cmd/skyman.go
	
ifeq ("$(UPX)", "")
	echo "upx not install"
else
	$(UPX) -q dist/skyman > /dev/null
endif

build-rpm: dist/skyman
	cp release/skyman.spec $(BUILD_SPEC)
	sed -i "s|VERSION|$(VERSION)|g" $(BUILD_SPEC)
	mkdir -p /root/rpmbuild/SOURCES
	cp dist/skyman etc/skyman-template.yaml locale/* static/* /root/rpmbuild/SOURCES
	rpmbuild -bb $(BUILD_SPEC)
	mv /root/rpmbuild/RPMS/x86_64/skyman-$(VERSION)-1.x86_64.rpm dist/
	rm -rf $(BUILD_SPEC)
