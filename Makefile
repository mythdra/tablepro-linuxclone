.PHONY: dev build build-all test lint clean up down

dev:
	wails dev

build:
	wails build

build-all:
	wails build -platform darwin
	wails build -platform windows
	wails build -platform linux

test:
	go test -v ./...
	cd frontend && npm test

lint:
	go vet ./...
	cd frontend && npm run lint

clean:
	rm -rf build/
	go clean
	cd frontend && rm -rf node_modules/ dist/

up:
	docker-compose up -d

down:
	docker-compose down