FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY front/ .
RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build

FROM golang:1.25-alpine AS backend-builder

ADD go.mod go.sum /
RUN go mod download

WORKDIR /app
COPY . .
RUN pwd
RUN ls -la
RUN ls -la ./cmd
RUN ls -la ./cmd/lightcall/
RUN CGO_ENABLED=0 go build -o lightcall ./cmd/lightcall/main.go

FROM alpine:latest
ENV TZ=Asia/Shanghai

RUN apk update
RUN apk add tzdata && cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone

WORKDIR /app
COPY --from=backend-builder /app/lightcall .
COPY --from=frontend-builder /app/dist ./public

EXPOSE 8090
CMD ["./lightcall", "serve", "--http=0.0.0.0:8090"]
