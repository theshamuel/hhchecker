VER=$(shell date +%Y-%m-%d-%H%M%S)
IMG_VER=
image-dev:
	docker build --build-arg VER=${VER} --build-arg SKIP_TEST=true --build-arg SKIP_LINTER=true -t ghcr.io/theshamuel/hhchecker .

image-prod:
	docker build --build-arg VER=${VER} -t ghcr.io/theshamuel/hhchecker:${IMG_VER} .

.PHONY: image-prod
