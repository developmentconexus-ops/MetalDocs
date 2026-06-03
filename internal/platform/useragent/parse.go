// Package useragent renders short, human-friendly device labels from raw
// User-Agent strings. Best-effort heuristics — no external dependency, no
// UA fingerprinting. Output is used in the Sessions & Security admin tab
// (SessionItem.deviceLabel) so admins can recognise which device a session
// belongs to without parsing UA strings by hand.
package useragent

import "strings"

const Unknown = "Unknown device"

// Label returns a short "<Browser> on <OS>" description for ua. Falls back to
// Unknown when the input is empty or unrecognised. Order of checks matters
// because UA strings frequently embed multiple tokens (e.g. Chrome UA
// contains "Safari" + "AppleWebKit").
func Label(ua string) string {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return Unknown
	}
	browser := detectBrowser(ua)
	os := detectOS(ua)
	if browser == "" && os == "" {
		return Unknown
	}
	if browser == "" {
		return os
	}
	if os == "" {
		return browser
	}
	return browser + " on " + os
}

func detectBrowser(ua string) string {
	switch {
	case strings.Contains(ua, "Edg/"), strings.Contains(ua, "Edge/"):
		return "Edge"
	case strings.Contains(ua, "OPR/"), strings.Contains(ua, "Opera"):
		return "Opera"
	case strings.Contains(ua, "Firefox/"):
		return "Firefox"
	case strings.Contains(ua, "Chrome/"), strings.Contains(ua, "CriOS/"):
		// CriOS = Chrome on iOS. Treat as Chrome — the OS check below
		// tags it as iOS so the label still reads correctly.
		return "Chrome"
	case strings.Contains(ua, "FxiOS/"):
		return "Firefox"
	case strings.Contains(ua, "Safari/") && strings.Contains(ua, "Mobile/"):
		return "Mobile Safari"
	case strings.Contains(ua, "Safari/"):
		return "Safari"
	case strings.Contains(ua, "curl/"):
		return "curl"
	case strings.Contains(ua, "PostmanRuntime"):
		return "Postman"
	case strings.Contains(ua, "Go-http-client"):
		return "Go HTTP client"
	}
	return ""
}

func detectOS(ua string) string {
	switch {
	case strings.Contains(ua, "Windows NT"):
		return "Windows"
	case strings.Contains(ua, "Android"):
		return "Android"
	case strings.Contains(ua, "iPhone"), strings.Contains(ua, "iPad"), strings.Contains(ua, "iPod"):
		return "iOS"
	case strings.Contains(ua, "Mac OS X"), strings.Contains(ua, "Macintosh"):
		return "macOS"
	case strings.Contains(ua, "CrOS"):
		return "ChromeOS"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	}
	return ""
}
