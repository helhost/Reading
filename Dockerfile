FROM golang:1.25 AS server-build
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN go build -o server .

# --- build client ---
FROM node:20-alpine AS client-build
WORKDIR /client
COPY client/ .
RUN mkdir -p /dist && cp -r . /dist

# --- runtime ---
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=server-build /app/server .
COPY --from=client-build /dist ./client
RUN mkdir -p /data
ENV DB_PATH=/data/data.db
ENV ALLOW_ORIGIN=*
EXPOSE 8080
CMD ["./server"]
