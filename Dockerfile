FROM golang:1.23 AS builder

WORKDIR /workspace

COPY . .

RUN go build -o /workspace/actioneer ./cmd/main.go

FROM golang:1.23

WORKDIR /workspace

COPY --from=builder /workspace/actioneer /workspace/actioneer
COPY --from=builder /workspace/.actioneer/** /workspace/.actioneer
COPY .actioneer /workspace/.actioneer

ENTRYPOINT ["/workspace/actioneer"]
