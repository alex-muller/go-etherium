mkdir -p build
mkdir -p build/bindings

files=`ls ./contracts/`

for entry in $files
do
  name="${entry%.*}"
  solc --abi ./contracts/$entry -o ./build/abi --overwrite
  solc --bin ./contracts/$entry -o ./build/bin/

  abigen --abi=./build/abi/$name.abi --pkg bindings --type $name --out ./build/bindings/$name.go --bin ./build/bin/$name.bin
done

chmod -R a=rwX build