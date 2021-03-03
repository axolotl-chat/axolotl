CARGO_VERSION := $(shell cargo --version 2>/dev/null)
GO_VERSION := $(shell go version 2>/dev/null)
GIT_VERSION := $(shell git --version 2>/dev/null)
uname_p := $(shell uname -p)
uname_s := $(shell uname -s)

all:
ifdef CARGO_VERSION
	@echo "Found cargo with version $(CARGO_VERSION)"
else
	@echo cargo not found, please install cargo and rust
	exit 1
endif
ifdef GO_VERSION
	@echo "Found go with version $(GO_VERSION)"
else
	@echo go not found, please install go
	exit 1
endif
ifdef GIT_VERSION
	@echo "Found go with version $(GIT_VERSION)"
else
	@echo go not found, please install git
	exit 1
endif
ifeq ($(uname_s), Linux)
ifeq ($(shell uname --hardware-platform), x86_64)
	@echo "get zkgroup $(PLATFORM)"
	go get -d github.com/nanu-c/zkgroup
	&& git submodule update \
	&& cd lib/zkgroup \
	&& cargo build --release --verbose
	mv libzkgroup.so libzkgroup_linux_amd64.so
else ifeq ($(shell uname --hardware-platform), aarch64)
	@echo "get zkgroup $(PLATFORM)"
	go get -d github.com/nanu-c/zkgroup
	&& git submodule update \
	&& cd lib/zkgroup \
	&& cargo build --release --verbose
	mv libzkgroup.so libzkgroup_linux_arm64.so
else 
	@echo architecture not supported
	exit 1
endif
else 
	@echo "platform not supported $(uname_s)"
	exit 1
endif

copy-lib:
ifdef GO_VERSION
	@echo "Found go with version $(GO_VERSION)"
else
	@echo go not found, please install go
	exit 1
endif
ifeq ($(uname_s), Linux)
ifeq ($(shell uname --hardware-platform), x86_64)
	@echo "get zkgroup $(PLATFORM)"
	go get -d github.com/nanu-c/zkgroup
else ifeq ($(shell uname --hardware-platform), aarch64)
	@echo "get zkgroup $(PLATFORM)"
	go get -d github.com/nanu-c/zkgroup
else 
	@echo architecture not supported
	exit 1
endif
else 
	@echo "platform not supported $(uname_s)"
	exit 1
endif

install-zkgroup:
ifeq ($(uname_s), Linux)
ifeq ($(shell uname --hardware-platform), x86_64)
	@echo "install libzkgroup to /usr/lib"
	cp ./libzkgroup_linux_amd64.so /usr/lib/
else ifeq ($(shell uname --hardware-platform), aarch64)
	@echo "install libzkgroup to /usr/lib"
	cp ./libzkgroup_linux_arm64.so  /usr/lib/

else 
	@echo architecture not supported
	exit 1
endif
else 
	@echo "platform not supported $(uname_s)"
	exit 1
endif

run:
ifdef GO_VERSION
	@echo "Found go with version $(GO_VERSION)"
	LD_LIBRARY_PATH=$(PWD) go run .
else
	@echo go not found, please install go
	exit 1
endif
