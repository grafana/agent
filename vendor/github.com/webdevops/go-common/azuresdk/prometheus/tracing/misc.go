package tracing

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func checkIfEnvVarIsEnabled(name string, defaultVal bool) bool {
	status := defaultVal

	val := os.Getenv(name)
	val = strings.ToLower(strings.TrimSpace(val))

	switch val {
	case "1", "true", "y", "yes", "enabled":
		status = true

	case "0", "false", "n", "no", "disabled":
		status = false
	}

	return status
}

func checkIfEnvVarContains(name string, value string, defaultVal bool) bool {
	envVal := strings.TrimSpace(os.Getenv(name))

	if envVal != "" {
		for _, part := range envVarSplit.Split(envVal, -1) {
			if strings.EqualFold(part, value) {
				return true
			}
		}

		return false
	}

	return defaultVal
}

func extractTenantIdFromRequest(r *http.Response) string {
	authToken := r.Request.Header.Get("authorization")
	if strings.HasPrefix(authToken, "Bearer") {
		authToken = strings.TrimSpace(strings.TrimPrefix(authToken, "Bearer"))
		authTokenParts := strings.Split(authToken, ".")
		if len(authTokenParts) == 3 {
			if val, err := base64.RawURLEncoding.DecodeString(authTokenParts[1]); err == nil {
				jwt := azureJwtAuthToken{}
				if err := json.Unmarshal(val, &jwt); err == nil {
					return jwt.Tid
				}
			}
		}
	}

	return ""
}
