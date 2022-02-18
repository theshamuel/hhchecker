FROM ghcr.io/theshamuel/baseimg-go-build:1.16.4 as builder

ARG VER
ARG SKIP_TEST
ARG SKIP_LINTER
ENV GOFLAGS="-mod=vendor"

LABEL org.opencontainers.image.source https://github.com/theshamuel/hhchecker

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
    if [ -z "$VER" ] ; then \
    version="test" && \
    echo "version=$version"; \
    else version=${VER}; fi && \
    go build -o hhchecker -ldflags "-X main.version=$version -s -w" .

FROM ghcr.io/theshamuel/baseimg-go-app:1.0-alpine3.13

WORKDIR /srv
COPY --from=builder /build/app/hhchecker /srv/hhchecker

RUN chown -R appuser:appuser /srv && date
USER appuser

CMD [ "/srv/hhchecker" ]