# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="anderson.gomes.c@gmail.com"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Copy the vendor directory
COPY vendor/ ./vendor/

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Add wait-for-it.sh script
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

COPY init-mysql.sh /init-mysql.sh
RUN chmod +x /init-mysql.sh

# Set the Current Working Directory inside the container to cmd/ordersystem
WORKDIR /app/cmd/ordersystem

# Export necessary ports
EXPOSE 8080
EXPOSE 8000
EXPOSE 50051

# Command to run the executable with wait-for-it script
CMD ["/wait-for-it.sh", "rabbitmq:5672", "--", "/init-mysql.sh", "&&", "go", "run", "-mod=vendor", "main.go", "wire_gen.go"]
