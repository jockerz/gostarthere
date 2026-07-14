run:
	# go run api/api.go api
	/home/sagala/go/bin/air

build:
	go build main.go

clean-packages: ## Clean packages
	go clean -modcache

requirements: ## Generate go.mod & go.sum files
	go mod tidy

test:
	go test ./...
	# go test -v ./...

smtp-test:
	python3 -m smtpd -n -c DebuggingServer localhost:1578

worker:
	./tmp/main worker

monitor-asynq:
	./asynqmon -redis-db 5 -redis-password RedisPassword -port 8081
