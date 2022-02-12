FROM theshamuel/baseimg-go-build:1.0 as builder

ARG VER
ARG SKIP_TEST
ARG SKIP_LINTER
ENV GOFLAGS="-mod=vendor"

RUN apk --no-cache add tzdata

ADD . /build/
WORKDIR /build/

#test
RUN \
    if [ -z "$SKIP_TEST" ]; then \
    go test -timeout=30s ./...; fi

#linter GolangCI
RUN \
    if [ -z "$SKIP_LINTER" ]; then \
    golangci-lint run --config .golangci.yml ; fi


RUN \
    if [ -z "$VER" ] ; then \
    version="test" && \
    echo "version=$version"; \
    else version=${VER}; fi && \
    go build -o hhchecker -ldflags "-X main.version=$version -s -w" .

FROM theshamuel/baseimg-go-app:latest

WORKDIR /srv
COPY --from=builder /build/hhchecker /srv/hhchecker

RUN chown -R appuser:appuser /srv && date
USER appuser

CMD [ "/srv/hhchecker" ]