// Package devianter is a client for DeviantArt's internal "_puppy" API, the
// JSON backend that deviantart.com's own web frontend calls.
//
// This is not the official, documented DeviantArt API. There is no application
// registration and no OAuth: the package authenticates the way a logged-out
// browser does, by fetching a guest session cookie and a CSRF token from the
// homepage. Everything reachable here is what an anonymous visitor can see.
// Because the endpoints are internal, DeviantArt can change or remove them
// without notice.
//
// # Usage
//
// Call [UpdateCSRF] once before anything else to establish the guest session.
// Every other call depends on the cookie and token it stores, and will fail
// until it has run:
//
//	if err := devianter.UpdateCSRF(); err != nil {
//		log.Fatal(err)
//	}
//
//	post, apiErr := devianter.GetDeviation("123456789", "someuser")
//	if apiErr.Reason != "" {
//		log.Fatal(apiErr.Error)
//	}
//	fmt.Println(post.Deviation.Title, post.IMG)
//
// The session does not refresh itself. A long-running program should call
// [UpdateCSRF] again when calls start failing, since tokens expire.
//
// # Errors
//
// Most functions return an [Error] value rather than a Go error. It is a struct,
// not an interface, so it is never nil; a call succeeded if Error.Reason is
// empty. Functions that can also fail on their arguments before any request is
// made (such as [PerformSearch] and [Group.Gallery]) return an ordinary error
// alongside it for that case.
//
// # Rate limiting and blocking
//
// DeviantArt sits behind CloudFront, which blocks IP addresses that request too
// aggressively. A blocked request surfaces as an [Error] whose Error field
// mentions CloudFront/WAF. This package does no rate limiting, retrying, or
// backoff of its own; a caller making bulk requests is expected to pace itself.
// Set [UserAgent] to identify your client and [Timeout] to bound each request.
package devianter
