BINARY_NAME=ms-chatpicpay-websocket-handler-api

build:
	go build -o bin/${BINARY_NAME} cmd/api/main.go
 
run: build
	./bin/${BINARY_NAME}

clean:
	go clean
	rm bin/${BINARY_NAME}

test:
	go test -race ./...

test_coverage:
	go test -short -race -coverprofile=cp.out ./...

dep:
	GOPRIVATE=github.com/PicPay go mod tidy && go mod download

fmt:
	go fmt ./...

vet:
	go vet ./...

docker-build:
	docker build . -f Dockerfile.dev -t ${BINARY_NAME} --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} --build-arg GITHUB_USER=${GITHUB_USER}

docker-run:
	docker run -p 8080:8080 ${BINARY_NAME}

govulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest --test ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2 run --out-format colored-line-number