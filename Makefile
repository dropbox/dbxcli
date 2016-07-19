VERSION=1.2.0
BINARY=dbxcli
all: clean binary
binary:
	go build -ldflags "-X main.version=${VERSION}" -o ${BINARY}
clean:
	rm -f dbxcli
install:
test:
