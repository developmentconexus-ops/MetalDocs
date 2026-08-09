package authn

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	authapp "metaldocs/internal/modules/auth/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/config"
)

// Enabled reports whether authentication is enabled, read from
// METALDOCS_AUTH_ENABLED (defaults to true when unset).
func Enabled() bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("METALDOCS_AUTH_ENABLED")))
	if raw == "" {
		return true
	}
	return raw == "1" || raw == "true" || raw == "yes" || raw == "on"
}

// CacheTTL returns the authz cache TTL, read from
// METALDOCS_AUTHZ_CACHE_TTL_SECONDS (defaults to 30 seconds when unset or invalid).
func CacheTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv("METALDOCS_AUTHZ_CACHE_TTL_SECONDS"))
	if raw == "" {
		return 30 * time.Second
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 30 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

// LoadRuntimeConfig reads the authn/authz runtime configuration from
// environment variables, enforcing that METALDOCS_AUTH_ENABLED=false is only
// permitted when APP_ENV=local.
func LoadRuntimeConfig() (authapp.Config, error) {
	rawAppEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if !Enabled() && rawAppEnv != "local" {
		return authapp.Config{}, errors.New("METALDOCS_AUTH_ENABLED=false only permitted when APP_ENV=local")
	}

	appEnv := rawAppEnv
	if appEnv == "" {
		appEnv = "local"
	}

	sessionSecret := strings.TrimSpace(os.Getenv("METALDOCS_AUTH_SESSION_SECRET"))
	if sessionSecret == "" && Enabled() {
		return authapp.Config{}, fmt.Errorf("METALDOCS_AUTH_SESSION_SECRET is required when auth is enabled")
	}

	sessionCookieName := parseSessionCookieName()

	sessionTTLHours, err := parseIntEnv("METALDOCS_AUTH_SESSION_TTL_HOURS", 12, func(v int) bool { return v >= 1 },
		"invalid METALDOCS_AUTH_SESSION_TTL_HOURS")
	if err != nil {
		return authapp.Config{}, err
	}

	// Sliding idle timeout. Defaults to 30 minutes (ISO 27001 path / backend
	// standardization parameters); override with METALDOCS_AUTH_SESSION_IDLE_MINUTES.
	// A value of 0 explicitly disables idle expiry (12h absolute TTL still applies).
	sessionIdleMinutes, err := parseIntEnv("METALDOCS_AUTH_SESSION_IDLE_MINUTES", 30, func(v int) bool { return v >= 0 },
		"METALDOCS_AUTH_SESSION_IDLE_MINUTES must be 0 or a positive integer")
	if err != nil {
		return authapp.Config{}, err
	}

	passwordMinLength, err := parseIntEnv("METALDOCS_AUTH_PASSWORD_MIN_LENGTH", 8, func(v int) bool { return v >= 8 },
		"invalid METALDOCS_AUTH_PASSWORD_MIN_LENGTH")
	if err != nil {
		return authapp.Config{}, err
	}

	maxFailedAttempts, err := parseIntEnv("METALDOCS_AUTH_LOGIN_MAX_FAILED_ATTEMPTS", 5, func(v int) bool { return v >= 3 },
		"invalid METALDOCS_AUTH_LOGIN_MAX_FAILED_ATTEMPTS")
	if err != nil {
		return authapp.Config{}, err
	}

	lockMinutes, err := parseIntEnv("METALDOCS_AUTH_LOGIN_LOCK_MINUTES", 15, func(v int) bool { return v >= 1 },
		"invalid METALDOCS_AUTH_LOGIN_LOCK_MINUTES")
	if err != nil {
		return authapp.Config{}, err
	}

	bootstrapEnabled, bootstrapUserID, bootstrapUsername, bootstrapName := parseBootstrapAdminIdentity(appEnv)

	trustedProxyCIDRs, err := config.LoadTrustedProxyCIDRs()
	if err != nil {
		return authapp.Config{}, err
	}

	cfg := authapp.Config{
		SessionCookieName:      sessionCookieName,
		SessionTTL:             time.Duration(sessionTTLHours) * time.Hour,
		SessionIdleTimeout:     time.Duration(sessionIdleMinutes) * time.Minute,
		SessionSecret:          authapp.Secret(sessionSecret),
		PasswordMinLength:      passwordMinLength,
		LoginMaxFailedAttempts: maxFailedAttempts,
		LoginLockDuration:      time.Duration(lockMinutes) * time.Minute,
		BootstrapAdminEnabled:  bootstrapEnabled,
		BootstrapAdminUserID:   bootstrapUserID,
		BootstrapAdminUsername: bootstrapUsername,
		BootstrapAdminEmail:    strings.TrimSpace(os.Getenv("METALDOCS_BOOTSTRAP_ADMIN_EMAIL")),
		BootstrapAdminPassword: authapp.Secret(os.Getenv("METALDOCS_BOOTSTRAP_ADMIN_PASSWORD")),
		BootstrapAdminName:     bootstrapName,
		CookieSecure:           config.ParseBoolEnv("METALDOCS_AUTH_COOKIE_SECURE", appEnv != "local"),
		TrustedOrigins:         splitCSV(os.Getenv("METALDOCS_AUTH_TRUSTED_ORIGINS")),
		OriginProtection:       config.ParseBoolEnv("METALDOCS_AUTH_ORIGIN_PROTECTION_ENABLED", Enabled()),
		TrustedProxyCIDRs:      trustedProxyCIDRs,
	}

	if cfg.BootstrapAdminEnabled && strings.TrimSpace(cfg.BootstrapAdminPassword.Value()) == "" {
		return authapp.Config{}, fmt.Errorf("METALDOCS_BOOTSTRAP_ADMIN_PASSWORD is required when bootstrap admin is enabled")
	}

	return cfg, nil
}

// parseSessionCookieName reads METALDOCS_AUTH_SESSION_COOKIE_NAME, defaulting
// to "metaldocs_session" when unset.
func parseSessionCookieName() string {
	name := strings.TrimSpace(os.Getenv("METALDOCS_AUTH_SESSION_COOKIE_NAME"))
	if name == "" {
		name = "metaldocs_session"
	}
	return name
}

// parseIntEnv reads an integer environment variable, returning defaultValue
// when unset. A present-but-unparsable value, or one that fails isValid, is
// reported via errMsg.
func parseIntEnv(name string, defaultValue int, isValid func(int) bool, errMsg string) (int, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || !isValid(parsed) {
		return 0, errors.New(errMsg)
	}
	return parsed, nil
}

// parseBootstrapAdminIdentity reads the bootstrap admin's enabled flag and
// identity fields, applying defaults for anything unset.
func parseBootstrapAdminIdentity(appEnv string) (enabled bool, userID, username, name string) {
	enabled = config.ParseBoolEnv("METALDOCS_BOOTSTRAP_ADMIN_ENABLED", appEnv == "local")
	userID = strings.TrimSpace(os.Getenv("METALDOCS_BOOTSTRAP_ADMIN_USER_ID"))
	if userID == "" {
		userID = "admin-local"
	}
	username = strings.TrimSpace(os.Getenv("METALDOCS_BOOTSTRAP_ADMIN_USERNAME"))
	if username == "" {
		username = "admin"
	}
	name = strings.TrimSpace(os.Getenv("METALDOCS_BOOTSTRAP_ADMIN_DISPLAY_NAME"))
	if name == "" {
		name = "Administrator"
	}
	return enabled, userID, username, name
}

// DevRoleMap parses METALDOCS_DEV_USER_ROLES as: user1:admin,user2:viewer|reviewer
func DevRoleMap() map[string][]iamdomain.Role {
	devRoleMapOnce.Do(func() {
		devRoleMapCached = parseDevRoleMap(strings.TrimSpace(os.Getenv("METALDOCS_DEV_USER_ROLES")))
	})
	return cloneDevRoleMap(devRoleMapCached)
}

var (
	devRoleMapOnce   sync.Once
	devRoleMapCached map[string][]iamdomain.Role
)

func parseDevRoleMap(raw string) map[string][]iamdomain.Role {
	if raw == "" {
		return map[string][]iamdomain.Role{"admin-local": {iamdomain.RoleSystemAdmin}}
	}

	out := map[string][]iamdomain.Role{}
	pairs := strings.Split(raw, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		userID := strings.TrimSpace(parts[0])
		rolesRaw := strings.TrimSpace(parts[1])
		if userID == "" || rolesRaw == "" {
			continue
		}
		roleParts := strings.Split(rolesRaw, "|")
		roles := make([]iamdomain.Role, 0, len(roleParts))
		for _, rp := range roleParts {
			r := strings.TrimSpace(rp)
			if r == "" {
				continue
			}
			roles = append(roles, iamdomain.Role(r))
		}
		if len(roles) > 0 {
			out[userID] = roles
		}
	}
	if len(out) == 0 {
		out["admin-local"] = []iamdomain.Role{iamdomain.RoleSystemAdmin}
	}
	return out
}

func cloneDevRoleMap(src map[string][]iamdomain.Role) map[string][]iamdomain.Role {
	out := make(map[string][]iamdomain.Role, len(src))
	for userID, roles := range src {
		out[userID] = append([]iamdomain.Role(nil), roles...)
	}
	return out
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			items = append(items, value)
		}
	}
	return items
}
