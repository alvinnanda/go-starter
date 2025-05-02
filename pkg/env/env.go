package env

import (
	"os"
	"strconv"
	"time"
)

// Get retrieves an environment variable or returns a default value if not set
func Get(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetInt retrieves an environment variable as an integer or returns a default value
func GetInt(key string, defaultValue int) int {
	strValue := Get(key, "")
	if strValue == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetBool retrieves an environment variable as a boolean or returns a default value
func GetBool(key string, defaultValue bool) bool {
	strValue := Get(key, "")
	if strValue == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(strValue)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// GetDuration parses a duration string from environment variable or returns a default value
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	strValue := Get(key, "")
	if strValue == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(strValue)
	if err != nil {
		return defaultValue
	}

	return duration
}
