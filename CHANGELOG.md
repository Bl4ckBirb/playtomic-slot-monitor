# Changelog

## [0.3.1] - 2026-07-25

### Fixed

- Point the client at `api.app.playtomic.io`. The old default, `api.playtomic.io`, is behind a CloudFront rule that 403s every request, which is why calls stopped working. The app host wants nothing but a valid token.
- Login now sends `requested_user_roles`, and refresh sends the token with the roles rather than a `grant_type`, matching what the app sends.

## [0.3.0] - 2025-06-14

### Added

- Courses (`SearchCourses`, `AllCourses`, `GetCourse`), tournaments (`SearchTournaments`, `AllTournaments`, `GetTournament`), the authenticated profile (`GetMe`, `GetUser`), social users and stats (`SearchSocialUsers`, `GetUserStats`), and `GetAuthMethods`.
- New search filters seen on the wire: date windows and `MatchStatus`/`UserID`/`PlayerUserID` on matches, sport and location on lessons, `WithProperties` on tenants, sport and course filters on classes.

### Changed

- `AvailabilityParams.TenantID string` becomes `TenantIDs []string`, so one call can cover several clubs. Wrap an existing single ID in a slice. `StartMin`/`StartMax` are added alongside the local-time window.

## [0.2.0] - 2025-05-31

Breaking. Pre-1.0, and the shape needed the change.

### Migration

| Before | After |
| --- | --- |
| `GetClasses`, `GetLessons`, `GetMatches` | `SearchClasses`, `SearchLessons`, `SearchMatches` |
| `*client.APIError` | `*client.Error`, plus sentinels for `errors.Is` |
| `WithDebug(bool)` | `WithLogger(*slog.Logger)` |
| `StartDate`, `EndDate`, `CreatedAt` as `string` | `models.Time` |
| `FromStartDate string` | `time.Time` |
| `DefaultBaseURL` ending in `/v1` | the host alone, versions live in the paths |
| `Lesson.ReservationIDs any` | `[]string` |

`apiErr.StatusCode == 429` becomes `errors.Is(err, client.ErrRateLimited)`. Comparing integers still works, `*Error` keeps `StatusCode`.

### Added

- Tenant search and lookup by ID, and court availability.
- Authentication: `Login`, `Refresh`, `TokenSource`, and the `WithToken`, `WithTokenSource` and `WithCredentials` options.
- `NewFromEnv` and the `PLAYTOMIC_*` variables, so nothing deployment-specific is compiled in. `WithHeader` and `WithAuthPaths` cover the same ground programmatically.
- `AllClasses`, `AllLessons`, `AllMatches` and `AllTenants`, iterating every page through `iter.Seq2`.
- `WithBackoff` for the retry window and its cap.
- Runnable examples and a package doc.

### Changed

- Retries cover 429 and 5xx, not only transport failures, and only on idempotent methods. A POST is never replayed.
- The backoff window doubles per attempt and lands in its upper half. `Retry-After` is honoured in full, and a wait longer than the cap returns the response with `Error.RetryAfter` set rather than retrying early.
- `Error` carries the method, URL, status, request ID, `Retry-After` and a flattened snippet of the body. HTML bodies lose their tags, so a response from something in front of the API still reads.
- `models.Time` parses the zoneless layout, RFC 3339 and date-only values, and marshals back in whatever it read.

### Fixed

- `ToURLValues` dropped `coordinate` when tenant IDs were set, and the two test cases covering it asserted opposite things, so the suite could not pass.
- The request was built once and replayed inside the retry loop, so a retried body was already read to EOF.
- `WithRetries` with a negative count left a nil response for the deferred close to panic on.
- The first retry waited `attempt * 500ms`, which at attempt zero is no wait.
- Any non-200 collapsed to "Unexpected response from API" with the body discarded.
- An empty 200 decoded to a zero value and returned as success, so a by-ID lookup could hand back a nil and no error.
- `WithTimeout` and `WithHTTPClient` fought over the same field, and which one held depended on the order they were passed.
- `PLAYTOMIC_HEADERS` could override an explicit `WithUserAgent`.
- The credentials source held a mutex across the login, so a waiting caller could not honour its own context.
- A rejected token was never dropped, leaving a revoked credential resending it forever.
- Pagination stopped on a short page, which truncated silently whenever the server capped the requested size.
- Matches omitted `page` when it was zero while classes and lessons sent it.

## [0.1.0] - 2025-05-09

First release. Classes, lessons and matches.
