.PHONY: compile
compile:
	docker run -v$$(pwd):/work goeth sh scripts/compile.sh

.PHONY: run
run:
	rm -rf ./build/dev-chain
	geth  --log.debug --dev --ws --ws.origins "https://app.tryethernal.com" --http --http.api eth,web3,net --http.corsdomain "https://remix.ethereum.org" --datadir ./build/dev-chain

.PHONY: build
build:
	docker build -t goeth .