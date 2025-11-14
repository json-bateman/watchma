## Docker Deployment

### Building and Running with Docker

Build the Docker image:
```bash
docker build -t watchma .
```

Run the container:
```bash
docker run -p 58008:58008 \
  -e JELLYFIN_API_KEY=your_api_key \
  -e JELLYFIN_BASE_URL=your_jellyfin_url \
  -e PORT=58008 \
  -e LOG_LEVEL=INFO \
  -v $(pwd)/data:/app/data \
  watchma
```

### Using Docker Compose

Create a `.env` file with your Jellyfin credentials:
```
JELLYFIN_API_KEY=your_api_key
JELLYFIN_BASE_URL=your_jellyfin_url
```

Run with docker-compose:
```bash
docker-compose up -d
```

### Deploying to Docker Hub

1. Tag your image:
```bash
docker tag watchma your-dockerhub-username/watchma:latest
```

2. Push to Docker Hub:
```bash
docker push your-dockerhub-username/watchma:latest
```

3. Pull and run on any machine:
```bash
docker pull your-dockerhub-username/watchma:latest
docker run -p 58008:58008 \
  -e JELLYFIN_API_KEY=your_api_key \
  -e JELLYFIN_BASE_URL=your_jellyfin_url \
  your-dockerhub-username/watchma:latest
```
