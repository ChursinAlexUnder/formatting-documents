# Use the official Go image as the base image
FROM golang:1.21.6

# Set the working directory in the container
WORKDIR /www

# Copy the application files into the working directory
COPY . /www

# Build the application
RUN go build -o main .

# Expose port 8080
EXPOSE 5000

# Define the entry point for the container
CMD ["./main"]
