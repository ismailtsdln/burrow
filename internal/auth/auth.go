package auth

// Authenticator defines the interface for user authentication.
type Authenticator interface {
	Authenticate(reason string) (bool, error)
}

// Current returns the best available authenticator for the platform.
func Current() Authenticator {
	return getPlatformAuthenticator()
}
