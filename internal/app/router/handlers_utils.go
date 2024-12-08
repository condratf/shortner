package router

import (
	"net/http"
)

func getUserIDFromCookie(r *http.Request) (*string, error) {
	cookie, err := r.Cookie(userCookieName)
	if err != nil || !isValidCookie(cookie) {
		return nil, err
	}
	return &cookie.Value, nil
}

func isValidCookie(cookie *http.Cookie) bool {
	// cookie validation logic here
	return cookie != nil && cookie.Value != ""
}
