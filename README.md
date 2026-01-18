## <p align="center">![Watchma](public/watchma-outlined.png)</p>

<a  target="_blank" href="https://data-star.dev/" ><img src="public/datastar-rocket.png" width="32"/></a> + <a target="_blank" href="https://templ.guide/"><img src="public/templ.svg" width="120"/></a> + <a target="_blank" href="https://tailwindcss.com/"><img src="public/tailwind.png" width="36"/></a>

### Why I made this
How many times have you thought "Man there's nothing to watch"? But then you actually have hundreds of movies at your disposal and you're simply overwhelmed by choices and your friends can't actually decide on what to watch so you just end up scrolling Jellyfin and arguing?

Ok maybe this is a specific me problem, but if it's not, that's why I made this site, hopefully to make the process of picking a movie more fun.

The voting system is influenced by the young CGP Grey. [Everyone Should Vote More Than Once](https://www.youtube.com/watch?v=orybDrUj4vA)

## Dev Setup

### For this to work you must have 5 things installed and in your `$PATH`
- [task](https://github.com/go-task/task?tab=readme-ov-file)
- [go](https://go.dev/doc/install)
- [air](https://github.com/air-verse/air)
- [templ](https://github.com/a-h/templ?tab=readme-ov-file)
- [tailwindcss](https://github.com/tailwindlabs/tailwindcss/) - Download from Releases

### Setting .env variables
Copy `.env.example` to `.env` and fill in your values:
```bash
cp .env.example .env
```

**Optional:** Configure Jellyfin credentials to fetch real movie data. Without these, the app will use dummy data for testing:
- `JELLYFIN_API_KEY`
- `JELLYFIN_BASE_URL`

**Optional:** Add `OPENAI_API_KEY` for AI-generated game messages. Uses ~$0.01 per 100 games.

See `.env.example` for all available configuration options including `PORT`, `LOG_LEVEL`, and `IS_DEV`.  

### Using Bruno (basically offline postman)
I have some of the api endpoints saved in the _bruno folder [install bruno](https://www.usebruno.com/) if you want to use these endpoints. Make sure to copy your `.env` file in the root of the `_bruno` directory so the endpoints can use the environment variables in their requests. Jellyfin requires an API key in each of it's requests. [here's](https://docs.usebruno.com/secrets-management/dotenv-file) some info on how to store bruno secrets.

### The Layers / Setup
I have this project split into distinct layers to keep things organized. 

- web/features: Handle HTTP Requests, perform handler operations for rendering pages, similar setup to [Northstar](https://github.com/zangster300/northstar)
- web/views: Just for common templ files / or pages
- pkg/: Handle all of the business logic. Grab data from external services with providers (i.e. Jellyfin/openAi)
- db: Handles the CRUD operations directly on the database. Migrations, sqlc, etc. 

### Testing

Right now I have basic clickthrough testing setup using playwright, `npm test` or `npm run test:ui` for detailed results.
Testing is put in place on push to prevent regression, if your git client is giving you an error, try running `npm test` to see more
detailed errors.

**NOTE:** You cannot have the app running when you run playwright tests, full e2e tests mean it needs the ports for NATS and the app itself

## Deployment

### Building and Pushing Docker Image
Build and push multi-platform Docker images (linux/amd64 and linux/arm64) to Docker Hub:

```bash
# Build and push latest
./build-and-push.sh

# Build and push specific version
./build-and-push.sh v1.2.3
```

**Note:** Make sure you're logged in to Docker Hub first:
```bash
docker login
```

To use your own Docker Hub username, set the `DOCKER_USERNAME` environment variable:
```bash
DOCKER_USERNAME=yourusername ./build-and-push.sh
```

## Support

If you *really* enjoy using Watchma, [I'm a simple man, I like money](https://ko-fi.com/jsonbateman). But mainly, thanks for trying it out!
