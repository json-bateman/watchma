### Datastar + Templ + Tailwind

Trying out a new stack, the point of the site is just going to be to grab info from 
Jellyfin on my localhost and display the information. 

### Getting started

This doesn't have hot module reloading (yet) so when you make changes you have to look in the 
justfile for commands, I use `just run` to compile the css, templ files and run the server

### For this to work you must have 3 things installed and in your `$PATH`
- [go](https://go.dev/doc/install)
- templ `go install github.com/a-h/templ/cmd/templ@latest`
- tailwindcss [The CLI tool is
here](https://github.com/tailwindlabs/tailwindcss/releases/tag/v4.1.8)
