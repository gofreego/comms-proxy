.PHONY: build run test proto migrate docker

build:
	go build -o application .
run:
	go run main.go
test:
	go test ./...
proto:
	sh ./api/protoc.sh
migrate:
	sql-migrator resources/configs/migrator.yaml
docker:
	docker build -t comms-proxy .
