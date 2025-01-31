gen:
	oapi-codegen \
		-package petstore \
		-generate echo-server,types \
		-o ./api/v1/generated.go \
		./petstore.yaml

make run_local:
	LOCAL=true go run main.go