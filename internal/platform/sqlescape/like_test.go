package sqlescape

import "testing"

func TestLikeEscape(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"plain", "plain"},
		{"50%", `50\%`},
		{"a_b", `a\_b`},
		{`a\b`, `a\\b`},
		{`100%_x\y`, `100\%\_x\\y`},
	}
	for _, c := range cases {
		if got := LikeEscape(c.in); got != c.want {
			t.Errorf("LikeEscape(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
