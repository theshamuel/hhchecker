VER=$(shell date +%Y-%m-%d-%H%M%S)
IMG_VER=
image-dev:
	docker build --build-arg VER=${VER} --build-arg SKIP_TEST=true --build-arg SKIP_LINTER=true -t ghcr.io/theshamuel/hhchecker .

image-prod:
	docker build --build-arg VER=${VER} -t ghcr.io/theshamuel/hhchecker:${IMG_VER} .

clean:
	- docker ps -a | grep -i "/bin/sh -c" | awk '{print $$1}' | xargs -n1 docker rm
	- docker images | grep -i "none" | awk '{print $$3}' | xargs -n1 docker rmi
	- docker rmi $$(docker images -q -f dangling=true)

.PHONY: image-prod
