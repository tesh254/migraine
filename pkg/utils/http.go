package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

func DownloadTemplate(urlStr string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to download template: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download template: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read template content: %v", err)
	}

	return string(body), nil
}

func ExtractSlugFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	fileName := path.Base(parsedURL.Path)

	slug := strings.TrimSuffix(fileName, path.Ext(fileName))

	return slug
}
