# Start from the latest golang base image
FROM ghcr.io/autamus/go:latest as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o hyprspace-relay .

# Start again with minimal envoirnment.
FROM ubuntu:latest

COPY --from=builder /app/hyprspace-relay /bin/hyprspace-relay

# Set the Current Working Directory inside the container
WORKDIR /app

CMD ["hyprspace-relay"]