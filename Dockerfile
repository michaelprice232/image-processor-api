FROM golang:1.20 as builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -ldflags="-w -s" -o image-processor-api ./cmd/processor/main.go


FROM gcr.io/distroless/base
COPY --from=builder /usr/src/app/image-processor-api /image-processor-api
ENTRYPOINT ["/image-processor-api"]