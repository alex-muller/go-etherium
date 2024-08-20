.PHONY: compile
compile:
	docker run -v$$(pwd):/work goeth sh scripts/compile.sh

.PHONY: run
run:
	geth --dev --http --http.api eth,web3,net --http.corsdomain "https://remix.ethereum.org" --datadir ./build/dev-chain

.PHONY: build
build:
	docker build -t goeth .