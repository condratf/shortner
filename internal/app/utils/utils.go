package utils

import "net/url"

func ConstructURL(baseURL string, paths ...string) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	for _, path := range paths {
		parsedURL.Path = parsedURL.Path + "/" + path
	}
	return parsedURL.String(), nil
}
