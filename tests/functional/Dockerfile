FROM python:3.13-slim-bullseye

# Set the working directory
WORKDIR /app

COPY requirements.txt .

# Install dependencies
RUN pip install -r requirements.txt

# Copy the test files into the container
COPY . .

# Command to run the tests
CMD ["pytest", "-v"]