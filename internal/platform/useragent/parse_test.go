package useragent

import "testing"

func TestLabel(t *testing.T) {
	cases := []struct {
		name string
		ua   string
		want string
	}{
		{"empty", "", Unknown},
		{"garbage", "definitely-not-a-real-ua", Unknown},
		{"chrome windows",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0 Safari/537.36",
			"Chrome on Windows"},
		{"firefox linux",
			"Mozilla/5.0 (X11; Linux x86_64; rv:127.0) Gecko/20100101 Firefox/127.0",
			"Firefox on Linux"},
		{"safari macos",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			"Safari on macOS"},
		{"mobile safari ios",
			"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Version/17.0 Mobile/15E148 Safari/604.1",
			"Mobile Safari on iOS"},
		{"chrome ios",
			"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1 (KHTML, like Gecko) CriOS/127.0 Mobile/15E148 Safari/604.1",
			"Chrome on iOS"},
		{"edge windows",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0 Safari/537.36 Edg/127.0",
			"Edge on Windows"},
		{"chrome android",
			"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0 Mobile Safari/537.36",
			"Chrome on Android"},
		{"curl", "curl/8.4.0", "curl"},
		{"os only", "Mozilla/5.0 (Windows NT 10.0)", "Windows"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Label(tc.ua); got != tc.want {
				t.Fatalf("Label(%q) = %q, want %q", tc.ua, got, tc.want)
			}
		})
	}
}
