package main

import (
	"io"
	"os"
	"time"
	"bytes"
	"regexp"
	"strings"
	"archive/zip"
	"path/filepath"
)

func DeleteOutput() {
	// Delete Output folder and Output.zip
	err := os.RemoveAll(outputPath)
	if err != nil {
		return
	}

	err = os.Remove(outputZip)
	if err != nil {
		return
	}
}

func ZipDirectory() error {
	// Create the Logs.zip file with contents from the Output directory
	var (
		buff      = new(bytes.Buffer)
		zipWriter = zip.NewWriter(buff)
		regex     = regexp.MustCompile(`Output\\[\w+]+\\`)
	)

	filepath.Walk(outputPath, func(file string, fi os.FileInfo, _ error) error {
		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(file)

		if !fi.IsDir() {
			if fi.Name() == "Logs.zip" { // Don't zip ourselves
				return nil
			}
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			defer data.Close()

			var folderPath = fi.Name()
			if regexS := regex.FindString(file); len(regexS) > 0 {
				folderPath = strings.Replace(regexS, "Output\\", "", -1) + fi.Name()
			}

			fileWriter, _ := zipWriter.Create(folderPath)
			if _, err := io.Copy(fileWriter, data); err != nil {
				return err
			}
		}

		return nil
	})

	if err := zipWriter.Close(); err != nil {
		return err
	}

	exportedZip, _ := os.Create(outputZip)
	defer exportedZip.Close()

	_, err := buff.WriteTo(exportedZip)
	if err != nil {
		return err
	}

	return nil
}

func CleanPath(filePath string) string {
	// Make sure filepath slashes do not collide
	return strings.ReplaceAll(filepath.Clean(filePath), `\`, `/`)
}

func ConvertUnixTime(nanoseconds int64) string {
	// Convert the passed integer value to a time.Time value and return it in its formatted state
	// `time` is the number of elapsed nanoseconds since the Unix epoch
	t := time.Unix(nanoseconds, 0)

	return t.Format(time.UnixDate)
}
