package html

import (
	"net/http"
)

// SecureCookie makes sure the passed cookies is only accessible
// to the browser, over HTTPS from the server's domain(for PUT, POST e.t.c)
func SecureCookie(appEnv string, cookie *http.Cookie) *http.Cookie {
	cookie.HttpOnly = true          // No JS access
	cookie.Secure = appEnv != "dev" // HTTPS only

	if appEnv != "dev" {
		cookie.SameSite = http.SameSiteLaxMode
	}

	return cookie
}

// LockCookie is SecureCookie with strict mode for same site settings
func LockCookie(appEnv string, cookie *http.Cookie) *http.Cookie {
	SecureCookie(appEnv, cookie)

	if appEnv != "dev" {
		cookie.SameSite = http.SameSiteStrictMode
	}

	return cookie
}
