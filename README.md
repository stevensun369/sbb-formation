# SBB formation forwarder

This Fiber API forwards train formation requests to OpenTransportData.swiss.

## Run locally

```sh
cp .env.example .env
# Set TOKEN in .env
go run .
```

The service listens on port `3001`.

Forward a request with:

```sh
curl 'http://localhost:3001/formation?date=2026-09-23&evu=BLSP&trainNumber=2806'
```

The forwarder maps `date` to the upstream API's `operationDate` parameter and
passes `evu` and `trainNumber` through unchanged. It authenticates upstream
requests with a Bearer token from `TOKEN`. The upstream response body, content
type, and status code are returned to the caller.

For a preformatted local query that automatically uses today's date and saves
the response as `response.json`, run:

```sh
./query.sh BLSP 2806
```

## Token configuration and build

At runtime, the application uses the compile-time token when one was embedded
with:

```sh
go build -ldflags "-X main.buildToken=$TOKEN"
```

If no token was embedded, it loads `TOKEN` from `.env`. The application exits
with an error if neither source provides a token.

`build.sh` reads `TOKEN` from `.env` and embeds it into statically linked
executables for Linux amd64, Windows amd64, and macOS amd64 and arm64 under
`dist/`.

```sh
cp .env.example .env
# Set TOKEN in .env
./build.sh
```
