gen:
	protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:. --grpc-gateway_out=. --openapiv2_out=:swagger
clean:
	rm pb/*.go
server:
	go run cmd/server/main.go -port=8080
client:
	go run cmd/client/main.go -addr="0.0.0.0:8080"
test:
	go test -cover -race ./...
cert:
	cd cert; bash ./gen.sh; cd ..

.PHONY: gen clean server client test cert
