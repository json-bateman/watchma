# Watchma

Stop arguing about what to watch. Watchma is an interactive movie voting system for groups. Connect to your Jellyfin server and let everyone vote and then Watchma movies hit yo screen. 🎬

The voting system is influenced by the young CGP Grey. [Everyone Should Vote More Than Once](https://www.youtube.com/watch?v=orybDrUj4vA)

## Setup: Docker Compose (Recommended)

Create a `docker-compose.yml` file:

```yaml
services:
  watchma:
    image: jsonbateman/watchma:latest
    container_name: watchma
    ports:
      - "58008:58008"
    volumes:
      - ./watchma-data:/data
    environment:
      - PORT=58008
      - LOG_LEVEL=INFO
      # Connect to your Jellyfin server
      - JELLYFIN_API_KEY=your_api_key_here
      - JELLYFIN_BASE_URL=https://jellyfin.example.com
      # Optional: Enable AI-powered game announcements
      - OPENAI_API_KEY=your_openai_key_here
    restart: unless-stopped
```

Then run:
```bash
docker compose up -d
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `58008` | Port the web interface runs on |
| `LOG_LEVEL` | No | `INFO` | Logging level: `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `JELLYFIN_API_KEY` | No | - | Your Jellyfin API key (uses dummy data if not provided) |
| `JELLYFIN_BASE_URL` | No | - | Your Jellyfin server URL (e.g., `https://jellyfin.example.com`) |
| `OPENAI_API_KEY` | No | - | OpenAI API key for AI-generated game messages (~$0.01 per 100 games) |

### Getting Your Jellyfin API Key

1. Log into your Jellyfin web interface
2. Go to **Dashboard** → **API Keys**
3. Click **+** to create a new key
4. Name it "Watchma" and copy the key


**Note:** Watchma works without Jellyfin. However, It'll use dummy movie data for testing. Jellyfin url and api key gives the full experience. 

### Volumes

| Container Path | Purpose | Required |
|----------------|---------|----------|
| `/data` | Database and persistent data | Yes |

**Important:** Make sure to mount `/data` to persist your users, game history, and settings.

## Unraid Installation

1. Go to the **Docker** tab
2. Click **Add Container**
3. Use template URL: `https://raw.githubusercontent.com/json-bateman/watchma/main/watchma.xml`

## Platform Support

This image supports multiple architectures:
- **linux/amd64** - Intel/AMD x86_64 (most PCs and servers)
- **linux/arm64** - ARM 64-bit (Raspberry Pi 4+, Apple Silicon)

Docker will automatically pull the correct architecture for your system.

## Usage

1. **Create a Room** - One person creates a room and shares the room name
2. **Draft Movies** - Everyone picks their favorite movies from your library
3. **Vote** - The system combines all drafts and everyone votes
4. **Watchma** - The winner is revealed with a fun announcement

## Healthcheck

Check if Watchma is running:
```bash
curl http://localhost:58008
```

Or check the logs:
```bash
docker logs watchma
```

## Upgrading

### With Docker Compose
```bash
docker compose pull
docker compose up -d
```

### With Docker CLI
```bash
docker pull jsonbateman/watchma:latest
docker stop watchma
docker rm watchma
# Then run your docker run command again
```

## Troubleshooting

### Can't Connect to Jellyfin
- Make sure your Jellyfin server is accessible from the Watchma container
- Verify your API key is correct
- Check `docker logs watchma` for connection errors
- If using HTTPS, ensure certificates are valid

### Database Locked Errors
Watchma uses SQLite. Make sure:
- Only one container is accessing the database
- The `/data` volume has proper permissions
- You're not running multiple instances pointing to the same data directory

## Links

- **GitHub:** [github.com/json-bateman/watchma](https://github.com/json-bateman/watchma)
- **Issues:** [Report a bug](https://github.com/json-bateman/watchma/issues)
- **Support:** [Ko-fi donations](https://ko-fi.com/jsonbateman) (if you really like it!)

## Example: Full Setup with Jellyfin

```bash
docker run -d \
  --name watchma \
  -p 58008:58008 \
  -v /path/to/watchma-data:/data \
  -e JELLYFIN_API_KEY="your_api_key_here" \
  -e JELLYFIN_BASE_URL="https://jellyfin.example.com" \
  -e OPENAI_API_KEY="sk-..." \
  -e LOG_LEVEL="INFO" \
  --restart unless-stopped \
  jsonbateman/watchma:latest
```

## Advanced: Using Specific Versions

Instead of `:latest`, you can pin to specific versions:

```yaml
image: jsonbateman/watchma:v1.0.0
```

Check [Docker Hub tags](https://hub.docker.com/r/jsonbateman/watchma/tags) for available versions.

---

**Enjoy!** If you find any issues or have suggestions, please [open an issue on GitHub](https://github.com/json-bateman/watchma/issues). 
