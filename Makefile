run:
	./mrunner ./workloads/fio/fio-file-4G.yaml

build-in-docker:
	docker run --rm -v "$$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.14 go build -v
