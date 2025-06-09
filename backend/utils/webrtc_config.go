package utils

import (
	"fmt"
	"os"
	"strings"
)

type WebRTCConfig struct {
	STUNServers []string
	TURNServers []TURNServerConfig
}

func LoadWebRTCConfig() *WebRTCConfig {
	return &WebRTCConfig{
		STUNServers: loadSTUNServers(),
		TURNServers: loadTURNServers(),
	}
}

func loadSTUNServers() []string {
	stunServers := getEnv("STUN_SERVERS", "stun:stun.l.google.com:19302,stun:stun1.l.google.com:19302")
	if stunServers == "" {
		return []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
			"stun:stun2.l.google.com:19302",
		}
	}
	return strings.Split(stunServers, ",")
}

func loadTURNServers() []TURNServerConfig {
	var turnServers []TURNServerConfig

	// Support multiple TURN servers with indexed environment variables
	for i := 1; i <= 5; i++ {
		urlKey := fmt.Sprintf("TURN_SERVER_%d_URL", i)
		userKey := fmt.Sprintf("TURN_SERVER_%d_USERNAME", i)
		credKey := fmt.Sprintf("TURN_SERVER_%d_CREDENTIAL", i)

		url := os.Getenv(urlKey)
		if url == "" {
			break // No more TURN servers
		}

		turnServer := TURNServerConfig{
			URL:        url,
			Username:   os.Getenv(userKey),
			Credential: os.Getenv(credKey),
		}

		turnServers = append(turnServers, turnServer)
	}

	return turnServers
}

// Default configuration for development
func GetDefaultWebRTCConfig() *WebRTCConfig {
	return &WebRTCConfig{
		STUNServers: []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
			"stun:stun2.l.google.com:19302",
			"stun:stun3.l.google.com:19302",
		},
		TURNServers: []TURNServerConfig{
			// Add your production TURN servers here
		},
	}
}
