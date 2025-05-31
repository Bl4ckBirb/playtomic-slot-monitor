# Endpoints

What the client covers, and the parameters each one accepts. `Search` returns a single page. `All` returns an iterator over every page.

## Classes

`GET /v1/classes` through `SearchClasses` and `AllClasses`.

`SearchClassesParams`: `Sort`, `Status`, `Type`, `TenantIDs`, `IncludeSummary`, `Size`, `Page`, `CourseVisibility`, `FromStartDate`, `Coordinate`, `Radius`.

```go
classes, err := c.SearchClasses(ctx, &models.SearchClassesParams{
	TenantIDs:      []string{tenantID},
	IncludeSummary: true,
	FromStartDate:  time.Now(),
})
```

## Lessons

`GET /v1/lessons` through `SearchLessons` and `AllLessons`.

`SearchLessonsParams`: `Sort`, `TenantID`, `TournamentVisibility`, `Status`, `Size`, `Page`, `FromStartDate`. This endpoint takes a single tenant ID rather than a list.

## Matches

`GET /v1/matches` through `SearchMatches` and `AllMatches`.

`SearchMatchesParams`: `Sort`, `HasPlayers`, `SportID`, `TenantIDs`, `Visibility`, `FromStartDate`, `Size`, `Page`.

## Tenants

`GET /v1/tenants` through `SearchTenants` and `AllTenants`, and `GET /v1/tenants/{id}` through `GetTenant`.

`SearchTenantsParams`: `TenantIDs`, `Name`, `SportID`, `PlaytomicStatus`, `Coordinate`, `Radius`, `Size`, `Page`.

Searching by coordinate is how you find the tenant IDs the other endpoints want.

```go
clubs, err := c.SearchTenants(ctx, &models.SearchTenantsParams{
	Coordinate: &models.Coordinate{Lat: 51.5074, Lon: -0.1278},
	Radius:     25000,
	SportID:    "PADEL",
})
```

## Availability

`GET /v1/availability` through `GetAvailability`.

`AvailabilityParams`: `TenantID`, `SportID`, `From`, `To`.

Slots come back grouped by resource, so one response covers the whole club.

```go
courts, err := c.GetAvailability(ctx, &models.AvailabilityParams{
	TenantID: tenantID,
	SportID:  "PADEL",
	From:     day,
	To:       day.Add(24 * time.Hour),
})
```

## Auth

`POST /v3/auth/login` through `Login`, `POST /v3/auth/token` through `Refresh`.

Both return a `Token`. Most callers want `WithCredentials` or `WithToken` instead and never touch these directly.
