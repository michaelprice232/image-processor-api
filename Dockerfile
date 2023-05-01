FROM golang:1.20 as builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=1 go build -v -o /usr/local/bin/image-processor-api ./cmd/processor/main.go


FROM scratch
COPY --from=builder /usr/local/bin/image-processor-api /image-processor-api
ENTRYPOINT ["/image-processor-api"]