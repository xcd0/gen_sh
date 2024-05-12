VERSION  := 0.0.1
REVISION := `git rev-parse --short HEAD`
FLAG     := -ldflags='-X main.version=$(VERSION) -X main.revision='$(REVISION)' -s -w -extldflags="-static" -buildid=' -a -tags netgo -installsuffix -trimpath 

all:
	cat ./makefile

fmt:
	goimports -w *.go
	gofmt -w *.go

generate:
	go generate # ../prebuild.sh‚ªŽÀs‚³‚ê‚éBgo generate‚Å‚È‚­‚Ä‚à‚æ‚¢B

build:
	rm -rf ./files
	make generate
	make fmt
	go build
	-@rm -rf embed.tar
	@echo
	-@make mv

release:
	rm -rf ./files
	make generate
	make fmt
	go build $(FLAG)
	make upx 
	-@rm -rf embed.tar
	@echo
	-@make mv
	@echo Success!
upx:
	upx --lzma ./gen_sh.exe
run:
	../gen_sh
version:
	-../gen_sh -v
	-../gen_sh --version
	-../gen_sh version
get:
	go mod tidy && go get
	go get github.com/alexflint/go-arg
	go get github.com/kdomanski/iso9660/util
	go get github.com/pkg/errors
	go get github.com/hjson/hjson-go/v4/hjson-cli
	go install golang.org/x/tools/cmd/...@latest
cp:
	-@cp -rf ./gen_sh.exe ..
mv:
	-@mv ./gen_sh.exe .. 2>/dev/null

