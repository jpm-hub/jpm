
build:
	mkdir -p out
	go build -o out/
	rm -f /Applications/cli/jpm
	cp -rf out/jpm /Applications/cli/jpm

