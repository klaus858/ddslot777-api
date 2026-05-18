FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN go build -o ddslot777-api .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/ddslot777-api ./ddslot777-api
ENV PORT=8080
EXPOSE 8080
CMD ["./ddslot777-api"]
