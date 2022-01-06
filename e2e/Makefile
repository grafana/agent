# We don't use -race here because there's races in docker/cli :)
test:
	go test -p=4 ./...

lint:
	GO111MODULE=on golangci-lint run -v --timeout=10m
