### Datastar + Templ + Tailwind

Trying out a new stack, the eventual goal for this site is to have a tournament style movie picker
that you can play by grabbing information directly from your jellyfin server. 

### Getting started

`just run` to compile the css, templ files and run the server.

For rebuilding download `air`, open 2 terminal windows, run `air` in one, and `just watch` in the
other, `air` will rebuild the go files on save, `just watch` will recompile the tailwind on save.

### For this to work you must have 5 things installed and in your `$PATH`
- [go](https://go.dev/doc/install)
- [just](https://github.com/casey/just?tab=readme-ov-file#installation)
- [air](https://github.com/air-verse/air)
- templ `go install github.com/a-h/templ/cmd/templ@latest`
- tailwindcss [The CLI tool is
here](https://github.com/tailwindlabs/tailwindcss/releases/tag/v4.1.8)

### NATS
Pub/Sub messaging system, for the chatroom functionality, you need to install nats and have it
running on its default port `:4222`

```bash
docker pull nats:latest
docker run -p 4222:4222 -ti nats:latest
```

### Using Bruno (basically offline postman)
I have all the api endpoints saved in the _bruno folder [install bruno](https://www.usebruno.com/)
if you want to use these endpoints. Make sure to copy your `.env` file in the root of the `_bruno` 
directory so the endpoints can use the environment variables in their requests.
[here's](https://docs.usebruno.com/secrets-management/dotenv-file) some info on how to store bruno
secrets.

### Understanding NATS

This is the flow for the app currently, the publish happens on the client, and then fans out to all
clients that are subscribed.

NATS Publish → Subscription Callback → Room History Storage → Client Channels → SSE Response → Browser Update

