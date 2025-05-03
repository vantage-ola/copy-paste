FROM golang:1.24.2 AS backend-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o copy-paste

FROM node:23 AS frontend-builder
WORKDIR /app
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM alpine:3.21 AS final
WORKDIR /app
COPY --from=backend-builder /app/copy-paste .
COPY --from=frontend-builder /app/dist ./dist
EXPOSE 8080
CMD ["./copy-paste"]
