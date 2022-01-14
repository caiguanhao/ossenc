bin:
	docker buildx build . \
		--build-arg VERSION=$$(git describe --tags --abbrev=0) \
		--target dist \
		--output dist/ \
		--platform=linux/amd64,linux/arm64,linux/arm/v7,windows/amd64

clean:
	rm -rf dist/

.PHONY: bin clean
