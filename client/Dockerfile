FROM node:20-alpine

# Install http-server globally
RUN npm install -g http-server

# Copy project into image
WORKDIR /app
COPY . .

# Expose port 80
EXPOSE 80

# Run http-server to serve index.html and src/
CMD ["http-server", "-p", "80", "-c-1", "."]

