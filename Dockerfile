# Stage 1: Build Go backend
FROM golang:1.22 as backend-build
WORKDIR /app
COPY . .
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o backend .

# Stage 2: Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=backend-build /app/backend /app/backend
EXPOSE 8080
CMD ["./backend"]
