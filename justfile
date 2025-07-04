# Ensure the templ binary is available
templ := "templ"
tailwindcss := "tailwindcss"

# Run templ code generation
tailwind:
    {{tailwindcss}} -i ./public/input.css -o ./public/style.css

generate: tailwind
    {{templ}} generate

run: generate
    go run ./cmd

watch:
    {{tailwindcss}} -i ./public/input.css -o ./public/style.css --watch
# Format Go and Templ files (optional bonus)
fmt:
    go fmt ./...
    {{templ}} fmt

dev:
    {{tailwindcss}} -i ./public/input.css -o ./public/style.css --watch

