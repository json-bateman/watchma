## <p align="center">![Watchma](public/watchma-outlined.png)</p>

<a  target="_blank" href="https://data-star.dev/" ><img src="public/datastar-rocket.png" width="32"/></a> + <a target="_blank" href="https://templ.guide/"><img src="public/templ.svg" width="120"/></a> + <a target="_blank" href="https://tailwindcss.com/"><img src="public/tailwind.png" width="36"/></a>

### Why I made this
How many times have you thought "Man there's nothing to watch"? But then you actually have hundreds of movies at your disposal and you're simply overwhelmed by choices and your friends can't actually decide on what to watch so you just end up scrolling Jellyfin and arguing?

Ok maybe this is a specific me problem, but if it's not, that's why I made this site, hopefully to make the process of picking a movie more fun, as of now it's a basic voting system with lobbies. But if I'm feeling ambitious I'd like to make this jackbox style one day.

## Dev Setup

### For this to work you must have 5 things installed and in your `$PATH`
- [task](https://github.com/go-task/task?tab=readme-ov-file)
- [go](https://go.dev/doc/install)
- [air](https://github.com/air-verse/air)
- [templ](https://github.com/a-h/templ?tab=readme-ov-file)
- [tailwindcss](https://github.com/tailwindlabs/tailwindcss/) - Download from Releases

### Setting .env variables
Must have a Jellyfin server with API key in your `.env` for this to run properly. For a fun message that plays before the end of the game, you can include an `OPENAI_API_KEY`, it only uses one small token request at the end of each game. 100 games has cost me ~ $0.01. 

`JELLYFIN_API_KEY`  
`JELLYFIN_BASE_URL`  
`OPENAI_API_KEY`
`PORT`  (to run the app on)  
`LOG_LEVEL` (DEBUG | INFO | WARN | ERROR)

### Using Bruno (basically offline postman)
I have some of the api endpoints saved in the _bruno folder [install bruno](https://www.usebruno.com/) if you want to use these endpoints. Make sure to copy your `.env` file in the root of the `_bruno` directory so the endpoints can use the environment variables in their requests. Jellyfin requires an API key in each of it's requests. [here's](https://docs.usebruno.com/secrets-management/dotenv-file) some info on how to store bruno secrets.

### The Layers / Setup
I have this project split into distinct layers to keep things organized. 

- web/features: Handle HTTP Requests, perform handler operations for rendering pages, similar setup to [Northstar](https://github.com/zangster300/northstar)
- web/views: Just for common templ files / or pages
- pkg/: Handle all of the business logic. Grab data from external services with providers (i.e. Jellyfin/openAi)
- db: Handles the CRUD operations directly on the database. Migrations, sqlc, etc. 

