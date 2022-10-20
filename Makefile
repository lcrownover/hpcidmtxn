all: server client

server:
	GOOS=linux GOARCH=amd64 go build -o bin/hpcidmtxn_server cmd/hpcidmtxn_server/main.go

client:
	GOOS=linux GOARCH=amd64 go build -o bin/hpcidmtxn_client cmd/hpcidmtxn_client/main.go

install: install_server install_client

install_server:
	cp bin/hpcidmtxn_server /usr/local/bin/

install_client:
	cp bin/hpcidmtxn_client /usr/local/bin/

clean:
	rm -f /usr/local/bin/hpcidmtxn_client
	rm -f /usr/local/bin/hpcidmtxn_server
