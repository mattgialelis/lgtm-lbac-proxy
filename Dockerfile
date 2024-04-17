# Stage 1: Build Vue.js frontend
FROM node:14 as frontend-build
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.20 as backend-build
WORKDIR /app
COPY . .
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o backend .

# Stage 3: Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=backend-build /app/backend /app/backend
COPY --from=frontend-build /app/dist /app/dist
EXPOSE 8080
CMD ["./backend"]
