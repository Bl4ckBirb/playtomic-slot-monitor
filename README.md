# go-playtomic-api

A Go client for the [Playtomic](https://playtomic.io) API, the padel and tennis court booking platform. No third-party dependencies.

## Status

Pre-1.0. Minor releases can break compatibility until 1.0.0. [CHANGELOG.md](./CHANGELOG.md) records what moved.

## Install

```bash
go get github.com/rafa-garcia/go-playtomic-api
```

## Quick start

```go
c := client.NewClient(client.WithTimeout(10 * time.Second))

matches, err := c.SearchMatches(ctx, &models.SearchMatchesParams{
	SportID:       "PADEL",
	TenantIDs:     []string{tenantID},
	FromStartDate: time.Now(),
	HasPlayers:    true,
})
if err != nil {
	log.Fatal(err)
}

for _, m := range matches {
	fmt.Printf("%s %s %.1f-%.1f\n", m.StartDate, m.Tenant.TenantName, m.MinLevel, m.MaxLevel)
}
```

## Endpoints

| Methods | Request |
| --- | --- |
| `SearchClasses`, `AllClasses` | `GET /v1/classes` |
| `SearchLessons`, `AllLessons` | `GET /v1/lessons` |
| `SearchMatches`, `AllMatches` | `GET /v1/matches` |
| `SearchTenants`, `AllTenants` | `GET /v1/tenants` |
| `GetTenant` | `GET /v1/tenants/{id}` |
| `GetAvailability` | `GET /v1/availability` |
| `Login` | `POST /v3/auth/login` |
| `Refresh` | `POST /v3/auth/token` |

`Search` returns one page. `All` walks every page.

## Configuration

```go
c := client.NewClient(
	client.WithBaseURL("https://api.playtomic.io"),
	client.WithTimeout(15*time.Second),
	client.WithRetries(3),
	client.WithBackoff(500*time.Millisecond, 10*time.Second),
	client.WithUserAgent("my-app/1.0"),
	client.WithHeader("X-Something", "value"),
	client.WithLogger(slog.Default()),
	client.WithHTTPClient(myHTTPClient),
)
```

Nothing deployment-specific is compiled in. `NewFromEnv` reads it instead, and options passed alongside still win:

```go
c, err := client.NewFromEnv()
```

| Variable | Effect |
| --- | --- |
| `PLAYTOMIC_BASE_URL` | API host |
| `PLAYTOMIC_USER_AGENT` | `User-Agent` header |
| `PLAYTOMIC_TIMEOUT` | Go duration, such as `15s` |
| `PLAYTOMIC_MAX_RETRIES` | Retry budget per request |
| `PLAYTOMIC_HEADERS` | Extra headers, one `Name: value` per line |
| `PLAYTOMIC_LOGIN_PATH`, `PLAYTOMIC_REFRESH_PATH` | Auth endpoints, if they move |
| `PLAYTOMIC_ACCESS_TOKEN` | Bearer token |
| `PLAYTOMIC_EMAIL`, `PLAYTOMIC_PASSWORD` | Credentials to log in with |

A variable that is set but unparseable is an error at construction rather than a quiet fall back to the default.

## Authentication

Three ways in, depending on where the token lives:

```go
client.WithToken(accessToken)              // one you already hold
client.WithCredentials(email, password)    // log in lazily, renew on expiry
client.WithTokenSource(mySource)           // your own storage
```

`WithCredentials` logs in on the first call that needs a token and refreshes it before expiry. Concurrent callers on a cold client produce one login between them, not one each.

`Login` and `Refresh` are exported if you would rather hold the token yourself. A rejected token is dropped, so a revoked credential recovers on the next call rather than wedging the client.

## Pagination

```go
for match, err := range c.AllMatches(ctx, params) {
	if err != nil {
		return err
	}
	fmt.Println(match.MatchID, match.StartDate)
}
```

The iterator stops when a page comes back empty, and breaking out of the loop stops the requests. Stopping on a short page would save one request, but the server is free to cap the size you ask for, and then every walk would quietly end after the first page. It works on a copy of your params, so your `Page` stays put.

## Errors

Status codes unwrap to sentinels, so branch with `errors.Is`:

```go
switch {
case errors.Is(err, client.ErrRateLimited):
	// apiErr.RetryAfter carries what the server asked for
case errors.Is(err, client.ErrUnauthorized):
	// token expired
case err != nil:
	var apiErr *client.Error
	if errors.As(err, &apiErr) {
		log.Printf("%d %s (request %s)", apiErr.StatusCode, apiErr.Message, apiErr.RequestID)
	}
}
```

`ErrBadRequest`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrRateLimited` and `ErrServer` are the set. `*client.Error` keeps the method, URL, status, request ID, `Retry-After` and a flattened snippet of the body, so a response that never reached the API still says what happened instead of failing on a JSON decode.

## Retries

Only idempotent methods are retried, on transport failures, 429 and 5xx, never
501. A POST is never replayed. The window doubles per attempt and lands in its
upper half, so a fleet of clients does not resynchronise on one outage.

A `Retry-After` is honoured in full rather than shortened. If the server asks for longer than the cap set by `WithBackoff`, the response comes back to you with `Error.RetryAfter` set instead of being retried early.

## Times

The API sends wall-clock timestamps with no zone. `models.Time` parses them, along with RFC 3339, date-only values and null, and a tenant's `Address.Timezone` is what turns one into a real instant.

## Contributing

Issues and pull requests welcome. Please include tests, and run `go test ./...` and `golangci-lint run` before opening one.

## Licence

MIT. See [LICENSE](./LICENSE).
