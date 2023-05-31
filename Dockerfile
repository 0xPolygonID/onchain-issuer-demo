##
## Build onchain-issuer-demo
##
FROM golang:1.20-bullseye as base

WORKDIR /build

COPY . .
RUN go mod download

RUN go build -o ./onchain main.go

# Build an onchain-issuer-demo image
FROM scratch

COPY ./onchain-issuer.settings.yaml /app/onchain-issuer.settings.yaml
COPY ./resolvers.settings.yaml /app/resolvers.settings.yaml
COPY --from=base /build/onchain /app/onchain
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

ENV HOST=0.0.0.0
ENV PORT=8080

# Command to run
ENTRYPOINT ["/app/onchain"]
