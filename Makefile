proto:
	docker run -it --rm -v $(shell pwd)/grpc:/src:rw -u $(shell id -u):$(shell id -g) -w /src namely/protoc-all -f all.proto -l go --go-source-relative -o .
	docker run -it --rm -v $(shell pwd)/components:/src:rw -u $(shell id -u):$(shell id -g) -w /src namely/protoc-all -f components.proto -l go --go-source-relative -o .

run_deps:
	docker-compose up -d

run_server:
	@go run ./cmd/server

run_client:
	@go run ./cmd/client --name Albert

run_bot:
	@go run ./cmd/bot
