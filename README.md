### Datastar + Templ + Tailwind

Trying out a new stack, the eventual goal for this site is to have a tournament style movie picker
that you can play by grabbing information directly from your jellyfin server. 

### Getting started

`task dev` to compile the css, templ files and run the server.
`task` to see all available commands

### For this to work you must have 5 things installed and in your `$PATH`
- [task](https://github.com/go-task/task?tab=readme-ov-file)
- [go](https://go.dev/doc/install)
- [air](https://github.com/air-verse/air)
- [templ](https://github.com/a-h/templ?tab=readme-ov-file)
- [tailwindcss](https://github.com/tailwindlabs/tailwindcss/) - Download from Releases

### Setting .env variables (`internal/config/settings.go`)
JELLYFIN_API_KEY=<Your Jellyfin API Key>
JELLYFIN_BASE_URL=<The URL where you have jellyfin configured> 
- (jellyfin's default port is http://localhost:8096)

PORT=<Your app port>
LOG_LEVEL=<DEBUG | INFO | WARN | ERROR>
SESSION_TIMEOUT=<user timeout i.e. 30m> (not implemented yet)

### Using Bruno (basically offline postman)
I have some of the api endpoints saved in the _bruno folder [install bruno](https://www.usebruno.com/)
if you want to use these endpoints. Make sure to copy your `.env` file in the root of the `_bruno` 
directory so the endpoints can use the environment variables in their requests. Jellyfin requires an
API key in each of it's requests.
[here's](https://docs.usebruno.com/secrets-management/dotenv-file) some info on how to store bruno
secrets.

