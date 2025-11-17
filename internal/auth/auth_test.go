package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateAndCheckHash(t *testing.T) {
	password1 := "password1"
	password2 := "password2"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			boolCheck, err := CheckPasswordHash(test.password, test.hash)
			if boolCheck == test.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestCreateAndCheckJWT(t *testing.T) {
	userUUID := uuid.New()
	expiresIn, _ := time.ParseDuration("2h")
	expiresNow, _ := time.ParseDuration("0s")
	password1 := "password1"
	password2 := "password2"
	jwt1, _ := MakeJWT(userUUID, password1, expiresIn)
	jwt2, _ := MakeJWT(userUUID, password2, expiresNow)

	tests := []struct {
		name     string
		password string
		jwt      string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			jwt:      jwt1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			jwt:      jwt1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different jwt",
			password: password2,
			jwt:      jwt1,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			jwt:      jwt1,
			wantErr:  true,
		},
		{
			name:     "Invalid jwt",
			password: password1,
			jwt:      "invalidjwt",
			wantErr:  true,
		},
		{
			name:     "Expired token",
			password: password2,
			jwt:      jwt2,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ValidateJWT(test.jwt, test.password)
			if (err != nil) != test.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
