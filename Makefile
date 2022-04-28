default: app

app:
	go build -o ./build/ ./cmd/app

mockgen:
	mockgen -source=./internal/app/domain/ports.go -destination=./internal/app/domain/tests/mocks/ports.go
	mockgen -source=./internal/app/domain/st_simple.go -destination=./internal/app/domain/tests/mocks/st_simple.go

tests: mockgen
	go test ./...

clean:
	rm -rf build