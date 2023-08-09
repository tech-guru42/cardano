FROM ghcr.io/blinklabs-io/go:1.19.12-1 AS build

WORKDIR /app
COPY . .
RUN make build

FROM cgr.dev/chainguard/glibc-dynamic
COPY --from=0 /app/cardano-node-api /bin/
ENTRYPOINT ["cardano-node-api"]