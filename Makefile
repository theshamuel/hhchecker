VER=$(shell date +%Y-%m-%d-%H%M%S)
image-dev:
	docker build --build-arg VER=${VER} --build-arg SKIP_TEST=true --build-arg SKIP_LINTER=true -t theshamuel/hhchecker .

image-prod:
	docker build --build-arg VER=${VER} -t theshamuel/hhchecker .

.PHONY: image-prod
