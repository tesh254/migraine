package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"regexp"
	"strings"
)

func FormatString(value string) string {
	value = strings.ReplaceAll(value, " ", "_")

	reg := regexp.MustCompile("[^a-zA-Z0-9_]+")

	value = reg.ReplaceAllString(value, "")

	return value
}

func StripText(text string) string {
	strippedString := strings.ReplaceAll(text, " ", "")
	strippedString = strings.ReplaceAll(strippedString, "\n", "")

	return strippedString
}

func GenerateChecksum(content string) string {
	md5Hash := md5.New()
	io.WriteString(md5Hash, content)
	hashBytes := md5Hash.Sum(nil)

	md5HashString := fmt.Sprintf("%x", hashBytes)

	return md5HashString
}
