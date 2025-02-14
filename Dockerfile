# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="anderson.gomes.c@gmail.com"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

#RUN go mod download

# Copy the vendor directory
COPY vendor/ ./vendor/

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Set the Current Working Directory inside the container to cmd/ordersystem
WORKDIR /app/cmd/ordersystem

# Export necessary ports
EXPOSE 8080
EXPOSE 8000
EXPOSE 50051

# Command to run the executable with wait-for-it script

# Command to run the executable
CMD ["go", "run", "-mod=vendor", "main.go", "wire_gen.go"]
