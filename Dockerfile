FROM ghcr.io/theshamuel/baseimg-go-build:1.20.1-1 as builder

ARG VER
ARG SKIP_TEST
ARG SKIP_LINTER
ENV GOFLAGS="-mod=vendor"

LABEL org.opencontainers.image.source='https://github.com/theshamuel/hhchecker'

RUN apk --no-cache add tzdata

ADD . /build/
WORKDIR /build/app

#test
RUN \
    if [ -z "$SKIP_TEST" ]; then \
    go test -timeout=30s ./...; fi

#linter GolangCI
RUN \
    if [ -z "$SKIP_LINTER" ]; then \
    golangci-lint run --config ../.golangci.yml ; fi


RUN \
    ref=$(git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD); \
    version=${ref}_$(git log -1 --format=%h)_$(date +%Y%m%dT%H:%M:%S); \
    if [ -n "$VER" ] ; then \
    version=${VER}_${version}; fi; \
    echo "version=$version"; \
    go build -o hhchecker -ldflags "-X main.version=$version -s -w" .

FROM ghcr.io/theshamuel/baseimg-go-app:1.0-alpine3.17

WORKDIR /srv
COPY --from=builder /build/app/hhchecker /srv/hhchecker

RUN chown -R appuser:appuser /srv && date
USER appuser

CMD [ "/srv/hhchecker" ]