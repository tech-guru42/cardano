FROM ghcr.io/blinklabs-io/go:1.21.1-2 AS build

WORKDIR /app
COPY . .
RUN make build

FROM cgr.dev/chainguard/glibc-dynamic
COPY --from=0 /app/cardano-node-api /bin/
USER root
ENTRYPOINT ["cardano-node-api"]
