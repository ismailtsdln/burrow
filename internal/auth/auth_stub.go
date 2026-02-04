//go:build !darwin

package auth

type stubAuthenticator struct{}

func (s *stubAuthenticator) Authenticate(reason string) (bool, error) {
	return true, nil // Always succeed on non-supported platforms
}

func getPlatformAuthenticator() Authenticator {
	return &stubAuthenticator{}
}
