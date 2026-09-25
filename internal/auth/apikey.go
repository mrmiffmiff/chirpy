package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("No Authorization Header present")
	}
	authHeaderSlice := strings.Split(authHeader, " ")
	if len(authHeaderSlice) != 2 || authHeaderSlice[0] != "ApiKey" {
		return "", fmt.Errorf("Incorrect Authorization Header Format")
	}
	return authHeaderSlice[1], nil
}
