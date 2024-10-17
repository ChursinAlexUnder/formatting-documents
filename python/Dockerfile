# # Use the official Python image as base (latest stable version)
# FROM python:3.12.0-slim

# # Set the working directory in the container
# WORKDIR /www

# # Copy the requirements file (ensure you have this in your repository root)
# COPY . /www

# # Install the Python dependencies
# RUN pip install --no-cache-dir -r requirements.txt

# # Expose the port that the app runs on
# EXPOSE 5000

# # Define the entry point for the container
# CMD ["waitress-serve", "--listen=0.0.0.0:5000", "python.editdocument:app"]


# Use the official Python image as the base image
FROM python:3.12

# Set the working directory in the container
WORKDIR /www/python

# Copy the application files into the working directory
COPY . /www

# Install the application dependencies
RUN pip install -r requirements.txt

# Set the FLASK_APP environment variable
ENV FLASK_APP=editdocument.py

# Define the entry point for the container
CMD ["flask", "run", "--host=0.0.0.0"]

