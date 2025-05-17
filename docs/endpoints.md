# Playtomic API Endpoints

This document provides a brief overview of the Playtomic API endpoints supported by this client.

## Classes

**Endpoint:** `/classes` **Client Method:** `SearchClasses`

Search for classes (academy sessions) with filtering options.

```go
// Example
params := &models.SearchClassesParams{
    TenantIDs:      []string{"tenant-id"},
    IncludeSummary: true,
    FromStartDate:  "2023-01-01T00:00:00",
}
classes, err := client.SearchClasses(ctx, params)
```

## Matches

**Endpoint:** `/matches` **Client Method:** `SearchMatches`

Search for matches with filtering options.

```go
// Example
params := &models.SearchMatchesParams{
    SportID:       "PADEL",
    TenantIDs:     []string{"tenant-id1", "tenant-id2"},
    FromStartDate: "2023-01-01T00:00:00",
}
matches, err := client.SearchMatches(ctx, params)
```

## Lessons

**Endpoint:** `/lessons` **Client Method:** `SearchLessons`

Search for lessons/tournaments with filtering options. Unlike the other endpoints, this one only accepts a single tenant ID.

```go
// Example
params := &models.SearchLessonsParams{
    TenantID:             "tenant-id", // Only a single tenant ID is supported
    TournamentVisibility: "PUBLIC",
    Status:               "REGISTRATION_OPEN,REGISTRATION_CLOSED",
    FromStartDate:        "2023-01-01T00:00:00",
}
lessons, err := client.SearchLessons(ctx, params)
```

## Model Conversion

When working with different player and tenant models:

```go
// Convert LessonPlayer to Player
player := models.LessonPlayerToPlayer(&lessonPlayer)

// Convert LessonTenant to Tenant
tenant := models.LessonTenantToTenant(&lessonTenant)
