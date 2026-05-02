.PHONY: dev build test lint clean

   dev:
   	air

   build: web-build
   	go build -o bin/defuse ./cmd/defuse

   web-build:
   	cd web && bun run build

   test:
   	go test ./...

   lint:
   	go vet ./...
   	gofmt -l . | (! grep .)

   clean:
   	rm -rf bin tmp assets/dist/* web/dist
