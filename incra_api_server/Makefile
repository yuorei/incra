gen:
	oapi-codegen \
		-package petstore \
		-generate echo-server,types \
		-o ./api/v1/generated.go \
		./petstore.yaml

fmt:
	gofmt -s -l -w .
	go mod tidy

dev:
	docker compose up
	