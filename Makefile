.PHONY: build_client
build_client:
	@go build -o esive_client ./cmd/client

.PHONY: proto
proto:
	docker run -it --rm -v $(shell pwd)/grpc:/src:rw -u $(shell id -u):$(shell id -g) -w /src namely/protoc-all -f all.proto -l go --go-source-relative -o .
	docker run -it --rm -v $(shell pwd)/components:/src:rw -u $(shell id -u):$(shell id -g) -w /src namely/protoc-all -f components.proto -l go --go-source-relative -o .
	docker run -it --rm -v $(shell pwd)/queue:/src:rw -u $(shell id -u):$(shell id -g) -w /src namely/protoc-all -f messages.proto -l go --go-source-relative -o .

.PHONY: run_deps
run_deps:
	docker-compose up -d

