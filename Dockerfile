# Use the official Go image as the base image
FROM golang:1.21.6

# Set the working directory in the container
WORKDIR /www

# Copy the application files into the working directory
COPY . /www

# Build the Go application
RUN go build -o main .

# Install Python and required packages
RUN apt-get update && apt-get install -y python python-pip
RUN pip3 install flask python-docx

# Expose ports for both Go and Python servers
EXPOSE 8080
EXPOSE 5000

# Start both servers
CMD ["sh", "-c", "python3 python/editdocument.py & ./main"]
