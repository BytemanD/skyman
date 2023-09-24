
UPX=$(shell which upx 2> /dev/null)
VERSION := $(shell go run cmd/skyman.go -v |awk '{{print $$3}}')

BUILD_SPEC=/tmp/skyman.spec

build:
	go mod download
	mkdir -p dist
	go build  -ldflags "-X main.Version=$(VERSION) -s -w" -o dist/ cmd/skyman.go
	
ifeq ("$(UPX)", "")
	echo "upx not install"
else
	$(UPX) -q dist/skyman > /dev/null
endif

build-rpm: dist/skyman
	cp release/skyman.spec $(BUILD_SPEC)
	sed -i "s|VERSION|$(VERSION)|g" $(BUILD_SPEC)
	mkdir -p /root/rpmbuild/SOURCES
	cp dist/skyman etc/skyman-template.yaml locale/* /root/rpmbuild/SOURCES
	rpmbuild -bb $(BUILD_SPEC)
	mv /root/rpmbuild/RPMS/x86_64/skyman-$(VERSION)-1.x86_64.rpm dist/
	rm -rf $(BUILD_SPEC)
