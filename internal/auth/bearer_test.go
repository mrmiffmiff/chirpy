package auth

import (
	"net/http"
	"testing"
)

func TestBearerTokens(t *testing.T) {
	tokenStringEx := "Properly-tested-string"
	headerExGood := make(http.Header)
	headerExBadShort := make(http.Header)
	headerExBadLong := make(http.Header)
	headerExBadNoBearer := make(http.Header)
	headerExBadNoAuth := make(http.Header)
	for _, h := range []*http.Header{&headerExGood, &headerExBadShort, &headerExBadLong, &headerExBadNoBearer, &headerExBadNoAuth} {
		h.Set("test-header-key", "test-header-val")
	}
	headerExGood.Set("Authorization", "Bearer "+tokenStringEx)
	headerExBadShort.Set("Authorization", "Bearer")
	headerExBadLong.Set("Authorization", "Bearer Test Test")
	headerExBadNoBearer.Set("Authorization", "Test Test")

	tests := []struct {
		name               string
		header             http.Header
		wantErr            bool
		matchedTokenString string
	}{
		{
			name:               "Token present in header",
			wantErr:            false,
			matchedTokenString: tokenStringEx,
			header:             headerExGood,
		},
		{
			name:               "Improperly-formatted auth header: Too short",
			wantErr:            true,
			matchedTokenString: "",
			header:             headerExBadShort,
		},
		{
			name:               "Improperly-formatted auth header: Too long",
			wantErr:            true,
			matchedTokenString: "",
			header:             headerExBadLong,
		},
		{
			name:               "Improperly-formatted auth header: Doesn't start with Bearer",
			wantErr:            true,
			matchedTokenString: "",
			header:             headerExBadNoBearer,
		},
		{
			name:               "No Auth Header",
			wantErr:            true,
			matchedTokenString: "",
			header:             headerExBadNoAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := GetBearerToken(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tokenString != tt.matchedTokenString {
				t.Errorf("GetBearerToken() expects %v, got %v", tt.matchedTokenString, tokenString)
			}
		})
	}
}
