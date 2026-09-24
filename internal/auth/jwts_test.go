package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateAndValidateTokens(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	goodSecret := "Correct Secret"
	badSecret := "Incorrect Secret"
	longDuration, _ := time.ParseDuration("1h")
	shortDuration, _ := time.ParseDuration("100ms")
	lastingToken, _ := MakeJWT(userID1, goodSecret, longDuration)
	expiringToken, _ := MakeJWT(userID2, goodSecret, shortDuration)

	time.Sleep(shortDuration)

	tests := []struct {
		name    string
		token   string
		id      uuid.UUID
		secret  string
		wantErr bool
		matchID bool
	}{
		{
			name:    "Testing basic token verification",
			token:   lastingToken,
			id:      userID1,
			secret:  goodSecret,
			wantErr: false,
			matchID: true,
		},
		{
			name:    "Wrong secret won't verify",
			token:   lastingToken,
			id:      userID1,
			secret:  badSecret,
			wantErr: true,
			matchID: false,
		},
		{
			name:    "Expired token rejected",
			token:   expiringToken,
			id:      userID2,
			secret:  goodSecret,
			wantErr: true,
			matchID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validation, err := ValidateJWT(tt.token, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && validation != tt.id {
				t.Errorf("ValidateJWT() expects %v, got %v", tt.id, validation)
			}
		})
	}
}
