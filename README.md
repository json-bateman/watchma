## <p align="center">![Watchma](public/watchma.png)</p>

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

### Setting .env variables (loaded by `internal/config/settings.go`)
Must have a Jellyfin server with API key in your `.env` for this to run properly.

`JELLYFIN_API_KEY`  
`JELLYFIN_BASE_URL`  
`PORT`  (to run the app on)  
`LOG_LEVEL` (DEBUG | INFO | WARN | ERROR)

### Running dev
`docker-compose up` to start the [NATS](https://nats.io/) Service.  
`task dev` to compile the css, templ files and run the server with Air and Tailwind in watch mode.  
`task` to see all available commands

### Using Bruno (basically offline postman)
I have some of the api endpoints saved in the _bruno folder [install bruno](https://www.usebruno.com/) if you want to use these endpoints. Make sure to copy your `.env` file in the root of the `_bruno` directory so the endpoints can use the environment variables in their requests. Jellyfin requires an API key in each of it's requests. [here's](https://docs.usebruno.com/secrets-management/dotenv-file) some info on how to store bruno secrets.

### The Layers / Setup
I have this project split into distinct layers to keep things organized. 

- handlers/: Handle HTTP Requests / Responses related to `watchma`.
- services/: Handle all of the business logic. Grab data from external services (i.e. Jellyfin)
- database/repository: Handles the CRUD operations directly on the database. 

- view/: Where all the `.templ` files live that do the actual rendering of the HTTP strings.
