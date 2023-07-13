# Use the official Golang image as the base image
FROM golang:1.17-alpine AS build

# Set the working directory outside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o app

# Use a minimal image for the final container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the build stage to the final container
COPY --from=build /app/app .
COPY --from=build /app/ipfs_cids.csv .

# Set the command to run the binary
CMD ["./app"]
