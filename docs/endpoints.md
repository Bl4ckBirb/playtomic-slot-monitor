# Endpoints

What the client covers, and the parameters each one accepts. `Search` returns a single page, `All` an iterator over every page, `Get` one thing by ID.

## Classes

`GET /v1/classes` through `SearchClasses` and `AllClasses`.

`SearchClassesParams`: `Sort`, `Status`, `Type`, `TenantIDs`, `IncludeSummary`, `Size`, `Page`, `CourseVisibility`, `FromStartDate`, `ToStartDate`, `SportID`, `CourseID`, `PlayerUserID`, `Coordinate`, `Radius`.

## Lessons

`GET /v1/lessons` through `SearchLessons` and `AllLessons`.

`SearchLessonsParams`: `Sort`, `TenantID`, `TournamentVisibility`, `Status`, `SportID`, `UserID`, `Coordinate`, `Radius`, `Size`, `Page`, `FromStartDate`, `ToStartDate`. This endpoint takes a single tenant ID rather than a list.

## Matches

`GET /v1/matches` through `SearchMatches` and `AllMatches`.

`SearchMatchesParams`: `Sort`, `HasPlayers`, `SportID`, `TenantIDs`, `Visibility`, `FromStartDate`, `ToStartDate`, `ToCreatedAt`, `MatchStatus`, `UserID`, `PlayerUserID`, `Size`, `Page`. `UserID: "me"` narrows to the authenticated player.

## Courses

`GET /v1/courses` through `SearchCourses` and `AllCourses`, and `GET /v1/courses/{id}` through `GetCourse`. A course is a series of scheduled classes.

`SearchCoursesParams`: `Sort`, `TenantID`, `Visibility`, `AvailablePlacesFrom`, `CourseEndsAfter`, `Coordinate`, `Radius`, `Size`, `Page`.

## Tournaments

`GET /v2/tournaments` through `SearchTournaments` and `AllTournaments`, and `GET /v2/tournaments/{id}` through `GetTournament`.

`SearchTournamentsParams`: `Sort`, `SportID`, `TenantID`, `Status`, `RegistrationStatus`, `Type`, `Visibility`, `UserID`, `FromStartDate`, `ToStartDate`, `AvailablePlaces`, `Coordinate`, `Radius`, `Size`, `Page`.

## Tenants

`GET /v1/tenants` through `SearchTenants` and `AllTenants`, and `GET /v1/tenants/{id}` through `GetTenant`.

`SearchTenantsParams`: `TenantIDs`, `Name`, `SportID`, `PlaytomicStatus`, `Coordinate`, `Radius`, `WithProperties`, `Size`, `Page`.

Searching by coordinate is how you find the tenant IDs the other endpoints want.

```go
clubs, err := c.SearchTenants(ctx, &models.SearchTenantsParams{
	Coordinate: &models.Coordinate{Lat: 51.5074, Lon: -0.1278},
	Radius:     25000,
	SportID:    "PADEL",
})
```

## Resources

`GET /v1/tenants/{id}/resources` through `GetResources`. Returns a club's courts
as `[]models.TenantResource` (`ResourceID`, `Name`, and typed `Properties`:
`ResourceType` indoor/outdoor, `ResourceSize` single/double, `ResourceFeature`).
Availability identifies courts only by `resource_id`, so this is how you resolve
those ids to names and indoor/outdoor.

```go
courts, err := c.GetResources(ctx, tenantID)
for _, court := range courts {
	fmt.Println(court.ResourceID, court.Name, court.IsIndoor())
}
```

## Availability

`GET /v1/availability` through `GetAvailability`.

`AvailabilityParams`: `TenantIDs`, `SportID`, `From`, `To`, `StartMin`, `StartMax`. `From` and `To` bound the local (wall-clock) start time, while `StartMin` and `StartMax` bound the absolute one. Pass more than one tenant to read several clubs in one call.

```go
courts, err := c.GetAvailability(ctx, &models.AvailabilityParams{
	TenantIDs: []string{tenantID},
	SportID:   "PADEL",
	From:      day,
	To:        day.Add(24 * time.Hour),
})
```

Slots come back grouped by resource, so one response covers the whole club.

## Users

`GET /v2/users/me` through `GetMe` returns the authenticated account. `GET /v2/users/{id}` through `GetUser` returns one by ID. Both need a token.

## Social

`GET /v1/social/users` through `SearchSocialUsers` returns players from the social graph. `GET /v1/social/users/{id}/stats` through `GetUserStats` returns follower counts, with `id` of `me` for the authenticated user.

`SearchSocialUsersParams`: `RequesterUserID`, `UserIDs`, `ExcludeFollowed`.

## Auth

`POST /v3/auth/login` through `Login`, `POST /v3/auth/token` through `Refresh`. Both return a `Token`, and most callers want `WithCredentials` or `WithToken` instead of calling them directly.

`GET /v3/auth/methods` through `GetAuthMethods` reports how an email can sign in. It sends no token, so you can call it before you have one.
