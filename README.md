# Cloud Drive (backend)

A personal Google Drive-style cloud storage app built to explore how file storage and authentication can be implemented from scratch.

## Tech Stack

**[Frontend](https://github.com/ethancut/cloud-drive-f)**
- [Astro](https://astro.build/)

**Backend**
- [Go](https://go.dev/)
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver
- [jwt](https://github.com/golang-jwt/jwt) - authentication
- [bcrypt](https://cs.opensource.google/go/x/crypto) - password hashing
- [godotenv](https://github.com/joho/godotenv) - environment variables injection
- [cors](https://github.com/rs/cors) - cross-origin request handling

## Credits

- [Cloud SVG icon](https://www.svgrepo.com/collection/dazzle-line-icons/) - Dazzle Line Icons collection by Dazzle UI
- [Download SVG icon](https://www.svgrepo.com/author/Solar%20Icons/) - Solar Icons


## DEV ENVIROMENT SETUP
 - Ensure you have the latest version of Go installed
 - Run `go mod tidy` to pull all dependencies to the project
 - Set the environment variables (see .env.template)
 - (Recommended) run the dev server using air for hot reload\
  (install with `go install github.com/air-verse/air@latest`)
    - If not using air, just do `go run cmd/api/main.go`