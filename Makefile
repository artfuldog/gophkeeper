GO_RUN_TEST_CMD=/usr/local/go/bin/go test

proto:			## Generate protobuf files
	@rm -f pb/*.go
	@protoc --go_out=. --go_opt=paths=import \
	--go-grpc_out=. --go-grpc_opt=paths=import \
	internal/proto/*.proto
	@protoc-go-inject-tag -input="./internal/pb/*.pb.go"

tests:			## Make relevent packages tests with clean cache
	@/usr/local/go/bin/go clean -testcache
	@echo "=== Package - internal/client ==="
	@$(GO_RUN_TEST_CMD) -cover ./internal/client/...
	@echo "\n=== Package - internal/common ==="
	@$(GO_RUN_TEST_CMD) -cover ./internal/common/...
	@echo "\n=== Package - internal/crypt ==="
	@$(GO_RUN_TEST_CMD) -cover ./internal/crypt/...
	@echo "\n=== Package - internal/logger ==="
	@$(GO_RUN_TEST_CMD) -cover ./internal/logger/...
	@echo "\n=== Package - internal/server ==="
	@$(GO_RUN_TEST_CMD) -cover ./internal/server/...

tests-all:		## Make all tests
	@/usr/local/go/bin/go clean -testcache
	@$(GO_RUN_TEST_CMD) -cover ./internal/...

tests-race:		## Make all tests with racing checking
	@/usr/local/go/bin/go clean -testcache
	@$(GO_RUN_TEST_CMD) -race -cover ./internal/...

bench:
	@$(GO_RUN_TEST_CMD) -run=Bench* ./internal/... -bench=. -benchtime=25000x -count=8 | grep Benchmark

mocks:			## Generate mocks for protobuf and database
	@mockgen -source=internal/pb/items_grpc.pb.go -destination=internal/mocks/mockgrpc/items.go -package=mockgrpc
	@mockgen -source=internal/pb/users_grpc.pb.go -destination=internal/mocks/mockgrpc/users.go -package=mockgrpc
	@mockgen -source=internal/server/db/db.go -destination=internal/mocks/mockdb/db.go -package=mockdb
	@mockgen -source=internal/server/authorizer/authorizer.go -destination=internal/mocks/mockauth/authorizer.go -package=mockauth

release-dry-run:
	@goreleaser build \
	&& goreleaser release --skip-publish --snapshow --rm-dist

release:
	@goreleaser build --rm-dist\
	&& goreleaser release --rm-dist

certs:			## Generate self-signed certificates to encrypt (./certs/)
	@mkdir -p certs
	@openssl genrsa -out certs/ca.key 4096
	@openssl req -new -x509 -key certs/ca.key -sha256 -subj "/C=RU/ST=MSK/O=GophKeeper" -days 365 -out certs/ca.crt
	@openssl genrsa -out certs/service.key 4096
	@openssl req -new -key certs/service.key -out certs/service.csr -config certificate.conf
	@openssl x509 -req -in certs/service.csr -CA certs/ca.cert -CAkey certs/ca.key -CAcreateserial \
		-out certs/service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext

certs-verify:		## Prints certificate for examination
	@openssl x509 -in certs/service.pem -text -noout

install-ca-cert:	## Install CA root certificate
	@sudo mkdir -p /etc/ssl/certs
	@sudo cp certs/ca.crt /etc/ssl/certs/gophkeeper.crt
	@sudo update-ca-certificates

run-server-race:	## Run server with race flag
	go run --race cmd/server/main.go -d 127.0.0.1:5432/gophkeeper \
	--db_user gksa -l debug -k 123456789f123456789q123456789pQ1 \
	-t 1800 -m 10000000 --tls-cert certs/service.pem --tls-key certs/service.key

run-server-race-notls:	## Run server with race flag and disaled tls
	go run --race cmd/server/main.go -d 127.0.0.1:5432/gophkeeper \
	--db_user gksa -l debug -k 123456789f123456789q123456789pQ1 \
	-t 1800 -m 10000000 --disable-tls

run-client-race:	## Run client with race flag
	go run --race cmd/client/main.go

run-client-race-notls:	## Run server with race flag and disaled tls
	go run --race cmd/client/main.go -t

help:	## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: help, list, proto, tests, tests-all, mocks, certs, certs-verify, install-ca-cert, release