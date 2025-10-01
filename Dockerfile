FROM golang:latest@sha256:2c89c41fb9efc3807029b59af69645867cfe978d2b877d475be0d72f6c6ce6f6 AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=1 go build -o /go/bin/app ./cmd/connector

FROM gcr.io/distroless/cc

COPY --from=build --chown=nonroot:nonroot /go/bin/app /
USER nonroot

ENTRYPOINT ["/app"]
