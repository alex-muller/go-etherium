.PHONY: compile
compile:
	docker run -v$$(pwd):/work goeth sh scripts/compile.sh

.PHONY: build
build:
	docker build -t goeth .