all: build run

build:
	docker build -t zoo . 
	
run:
	docker run -p 8000:8000 --rm -it zoo

test: python node go google

python:
	curl http://localhost:8000
	curl http://localhost:8000/a/b/c

node:
	curl http://localhost:8000/node
	curl http://localhost:8000/node/a/b/c

go:
	curl http://localhost:8000/go
	curl http://localhost:8000/go/a/b/c

google:
	@echo Click on "http://localhost:8000/google?q=abc"
