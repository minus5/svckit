version           := 0.0.35
docker_registry   := registry.dev.minus5.hr
docker_image_name := $(docker_registry)/amp_tester:v$(version)

.PHONY: help
help:
	@cat Makefile

.PHONY: build
build:
	mkdir -p ./build/bin ./build/app
	cp -r ./app/* ./build/app
	env GOOS=linux GOARCH=amd64 go build -o ./build/bin/amp_tester
	docker build -t $(docker_image_name) ./build
	docker push $(docker_image_name)

.PHONY: clean
clean:
	rm -rf ./build/bin
