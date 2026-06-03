// Package sqlescape provides helpers for safely embedding user-supplied
// values into Postgres LIKE/ILIKE patterns.
package sqlescape

import "strings"

// LikeEscape escapes Postgres LIKE pattern metacharacters (% and _) using
// the backslash escape character. The caller MUST append "ESCAPE '\\'" to
// the LIKE clause in the surrounding SQL so Postgres interprets the
// backslashes correctly.
func LikeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
