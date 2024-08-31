package html

import (
	"net/http"
)

// SecureCookie makes sure the passed cookies is only accessible
// to the browser, over HTTPS from the server's domain(for PUT, POST e.t.c)
func SecureCookie(isProd bool, cookie *http.Cookie) *http.Cookie {
	cookie.HttpOnly = true // No JS access
	cookie.Secure = isProd // HTTPS only

	if isProd {
		cookie.SameSite = http.SameSiteLaxMode
	}

	return cookie
}

// LockCookie is SecureCookie with stricter mode for same site settings
func LockCookie(isProd bool, cookie *http.Cookie) *http.Cookie {
	SecureCookie(isProd, cookie)

	if isProd {
		cookie.SameSite = http.SameSiteStrictMode
	}

	return cookie
}
