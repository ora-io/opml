SHELL := /bin/bash

build: submodules libunicorn mlvm contracts mlgo
.PHONY: build

submodules:
	# CI will checkout submodules on its own (and fails on these commands)
	if [[ -z "$$GITHUB_ENV" ]]; then \
		git submodule init; \
		git submodule update; \
	fi
.PHONY: submodules

# Approximation, use `make libunicorn_rebuild` to force.
unicorn/build: unicorn/CMakeLists.txt
	mkdir -p unicorn/build
	cd unicorn/build && cmake .. -DUNICORN_ARCH=mips -DCMAKE_BUILD_TYPE=Release
	# Not sure why, but the second invocation is needed for fresh installs on MacOS.
	if [ "$(shell uname)" == "Darwin" ]; then \
		cd unicorn/build && cmake .. -DUNICORN_ARCH=mips -DCMAKE_BUILD_TYPE=Release; \
	fi

# Rebuild whenever anything in the unicorn/ directory changes.
unicorn/build/libunicorn.so: unicorn/build unicorn
	cd unicorn/build && make -j8
	# The Go linker / runtime expects dynamic libraries in the unicorn/ dir.
	find ./unicorn/build -name "libunicorn.*" | xargs -L 1 -I {} cp {} ./unicorn/
	# Update timestamp on libunicorn.so to make it more recent than the build/ dir.
	# On Mac this will create a new empty file (dyn libraries are .dylib), but works
	# fine for the purpose of avoiding recompilation.
	touch unicorn/build/libunicorn.so

libunicorn: unicorn/build/libunicorn.so
.PHONY: libunicorn

libunicorn_rebuild:
	touch unicorn/CMakeLists.txt
	make libunicorn
.PHONY: libunicorn_rebuild


mlvm:
	cd mlvm && go build
.PHONY: mlvm

mlgo:
	cd mlgo && pip install -r requirements.txt
	cd examples/mnist_mips && ./build.sh
.PHONY: mlgo

contracts: nodejs
	npx hardhat compile
.PHONY: contracts

nodejs:
	if [ -x "$$(command -v pnpm)" ]; then \
		pnpm install; \
	else \
		npm install; \
	fi
.PHONY: nodejs

# Must be a definition and not a rule, otherwise it gets only called once and
# not before each test as we wish.
define clear_cache
	rm -rf /tmp/cannon
	mkdir -p /tmp/cannon
endef

clear_cache:
	$(call clear_cache)
.PHONY: clear_cache


test_contracts:
	$(call clear_cache)
	npx hardhat test
.PHONY: test_contracts


clean:
	rm -f minigeth/go-ethereum
	rm -f mipigo/minigeth
	rm -f mipigo/minigeth.bin
	rm -f mipsevm/mipsevm
	rm -rf artifacts
	rm -f unicorn/libunicorn.*
.PHONY: clean

mrproper: clean
	rm -rf cache
	rm -rf node_modules
	rm -rf mipigo/venv
	rm -rf unicorn/build
.PHONY:  mrproper
