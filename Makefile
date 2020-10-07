run: build-in-docker
	./mrunner run ./workloads/fio/fio-file-4G.yaml
	find results/ -name fio.txt -exec echo {} \; -exec cat {} \;

example: build-in-docker
	./mrunner template
	./mrunner run ./workloads/example/example.yaml
	find results/ -name result.json -exec echo {} \; -exec cat {} \;

build-in-docker:
	docker run --rm -v "$$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.14 go build -v
