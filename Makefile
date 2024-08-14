.PHONY: compile
compile:
	docker run -v$$(pwd):/work goeth sh scripts/compile.sh

.PHONY: run
run:
	docker run -p 8545:8545 goeth  geth --dev --http --http.api eth,web3,net --http.corsdomain "https://remix.ethereum.org"

.PHONY: build
build:
	docker build -t goeth .