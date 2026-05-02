FROM golang:1.26.2 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /build/api ./cmd/api


FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /build/api /api

ENTRYPOINT ["/api"]
CMD ["api"]
