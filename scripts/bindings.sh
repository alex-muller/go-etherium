mkdir -p build/bindings

files=`ls ./build/compile`

for entry in $files
do
  name="${entry%.*}"
  abigen --abi=./build/compile/$entry --pkg bindings --type $name --out ./build/bindings/$name.go
done

chmod -R a=rwX build