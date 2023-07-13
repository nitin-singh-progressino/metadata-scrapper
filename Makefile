.PHONY: build run clean docker-build docker-run docker-clean

# Build the Go binary
build:
	go build -o app

# Run the application
run:
	go run .

# Clean up build artifacts
clean:
	rm -f app

# Build the Docker image
docker-build:
	docker build -t mymicroservice .

# Run the Docker container
docker-run:
	docker run -p 8080:8080 mymicroservice

# Clean up the Docker image
docker-clean:
	docker rmi mymicroservice
