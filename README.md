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
