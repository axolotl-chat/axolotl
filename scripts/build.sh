go clean

echo "Build ut linux armhf"
clickable build

echo "Build linux amd64 snap"
mkdir -p build/linux-amd64-snap
snapcraft clean axolotl
snapcraft
mv *.snap build/linux-amd64-snap/

echo "Build linux amd64"
mkdir -p build/linux-amd64/axolotl-web
env GOOS=linux GOARCH=amd64 go build -o build/linux-amd64/axolotl .
cp axolotl-web/dist build/windows-amd64/axolotl-web -r

echo "Build linux arm5"
mkdir -p build/linux-arm5/axolotl-web
env GOOS=linux GOARCH=arm GOARM=5 go build -o build/linux-arm5/axolotl .
cp axolotl-web/dist build/linux-arm5/axolotl-web -r

echo "Build linux arm7"
mkdir -p build/linux-arm7/axolotl-web
env GOOS=linux GOARCH=arm GOARM=7 go build -o build/linux-arm7/axolotl .
cp axolotl-web/dist build/linux-arm7/axolotl-web -r

echo "Build linux arm64"
mkdir -p build/linux-arm64/axolotl-web
env GOOS=linux GOARCH=arm64 go build -o build/linux-arm64/axolotl .
cp axolotl-web/dist build/linux-arm64/axolotl-web -r

echo "Build windows amd64"
mkdir -p build/windows-amd64/axolotl-web
env GOOS=windows GOARCH=amd64 go build -o build/windows-amd64/axolotl.exe .
cp axolotl-web/dist build/windows-amd64/axolotl-web -r

echo "Build windows 386"
mkdir -p build/windows-386/axolotl-web
env GOOS=windows GOARCH=386 go build -o build/windows-386/axolotl.exe .
cp axolotl-web/dist build/windows-amd64/axolotl-web -r

echo "Build darwin amd64"
mkdir -p build/darwin-amd64/axolotl-web
env GOOS=darwin GOARCH=amd64 go build -o build/darwin-amd64/axolotl .
cp axolotl-web/dist build/darwin-amd64/axolotl-web -r
