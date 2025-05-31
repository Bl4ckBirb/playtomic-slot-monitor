// Package client is a Go client for the Playtomic API.
//
// A client is built with NewClient and functional options, or with NewFromEnv
// to take its settings from the PLAYTOMIC_* environment variables:
//
//	c := client.NewClient(
//		client.WithTimeout(15*time.Second),
//		client.WithCredentials(email, password),
//	)
//
// Search methods return one page. The matching All methods iterate every page:
//
//	for match, err := range c.AllMatches(ctx, params) {
//		if err != nil {
//			return err
//		}
//		fmt.Println(match.StartDate, match.Tenant.TenantName)
//	}
//
// Failures come back as *Error, which unwraps to a sentinel so errors.Is works
// on the status. The package has no third-party dependencies.
package client
