BINDIR = $(CURDIR)/../setup/bin
$(shell mkdir -p $(BINDIR))
mdm_svr="mdm_svr"

default:
	@echo $(BINDIR)

install:
	@GO111MODULE=on go mod download

build:
	@echo version: ${VERSION} date: ${DATE} os: linux-centOS
	@GO111MODULE=on GOOS=linux go build -o ../setup/bin/${mdm_svr} ./${mdm_svr}/src
	@echo build success!

clean:
	@echo '>>> clean build dir : $(BINDIR)'
	@rm -rf $(BINDIR)