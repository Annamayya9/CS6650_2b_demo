FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY product-service.go ./
RUN CGO_ENABLED=0 go build -trimpath -o /product-api product-service.go

FROM alpine:3.22
WORKDIR /app
COPY --from=build /product-api /app/product-api
USER 10001:10001
ENV GIN_MODE=release
EXPOSE 8080
ENTRYPOINT ["/app/product-api"]
