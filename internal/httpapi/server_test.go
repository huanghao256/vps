package httpapi

import "testing"

func TestValidBearerToken(t *testing.T) {
	t.Parallel()

	if !validBearerToken("Bearer secret-token", "secret-token") {
		t.Fatal("expected valid bearer token")
	}
	if validBearerToken("Bearer wrong-token", "secret-token") {
		t.Fatal("expected invalid bearer token")
	}
	if validBearerToken("", "secret-token") {
		t.Fatal("expected empty header to be invalid")
	}
}
