FROM golang:1.25 AS build

WORKDIR /app

# copy go module files first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# copy the rest of the source
COPY . .

# build binary
RUN go build -o server .

# -------- runtime stage --------
FROM debian:bookworm-slim

WORKDIR /app

# copy binary from build stage
COPY --from=build /app/server .

# app will create/use ./data.db inside container
EXPOSE 8080

CMD ["./server"]
