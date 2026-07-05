package application

import (
	"errors"
	"testing"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

func TestValidateOnboardTenantInput(t *testing.T) {
	valid := func() (name, slug, adminUserID, adminDisplayName, adminPassword string) {
		return "Acme Inc", "acme-inc", "admin@acme.test", "Acme Admin", "correct-horse-battery-staple"
	}

	t.Run("valid input passes", func(t *testing.T) {
		name, slug, adminUserID, adminDisplayName, adminPassword := valid()
		if err := validateOnboardTenantInput(name, slug, adminUserID, adminDisplayName, adminPassword); err != nil {
			t.Fatalf("validateOnboardTenantInput() = %v, want nil", err)
		}
	})

	cases := []struct {
		name   string
		mutate func(name, slug, adminUserID, adminDisplayName, adminPassword string) (string, string, string, string, string)
	}{
		{"blank name", func(_, slug, adminUserID, adminDisplayName, adminPassword string) (string, string, string, string, string) {
			return "", slug, adminUserID, adminDisplayName, adminPassword
		}},
		{"blank slug", func(name, _, adminUserID, adminDisplayName, adminPassword string) (string, string, string, string, string) {
			return name, "", adminUserID, adminDisplayName, adminPassword
		}},
		{"slug with uppercase", func(name, _, adminUserID, adminDisplayName, adminPassword string) (string, string, string, string, string) {
			return name, "Acme-Inc", adminUserID, adminDisplayName, adminPassword
		}},
		{"slug with underscore", func(name, _, adminUserID, adminDisplayName, adminPassword string) (string, string, string, string, string) {
			return name, "acme_inc", adminUserID, adminDisplayName, adminPassword
		}},
		{"blank admin user id", func(name, slug, _, adminDisplayName, adminPassword string) (string, string, string, string, string) {
			return name, slug, "", adminDisplayName, adminPassword
		}},
		{"blank admin display name", func(name, slug, adminUserID, _, adminPassword string) (string, string, string, string, string) {
			return name, slug, adminUserID, "", adminPassword
		}},
		{"short admin password", func(name, slug, adminUserID, adminDisplayName, _ string) (string, string, string, string, string) {
			return name, slug, adminUserID, adminDisplayName, "short"
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, slug, adminUserID, adminDisplayName, adminPassword := tc.mutate(valid())
			err := validateOnboardTenantInput(name, slug, adminUserID, adminDisplayName, adminPassword)
			if !errors.Is(err, iamdomain.ErrTenantOnboardValidation) {
				t.Fatalf("validateOnboardTenantInput() = %v, want iamdomain.ErrTenantOnboardValidation", err)
			}
		})
	}
}

func TestIsURLSafeSlug(t *testing.T) {
	cases := map[string]bool{
		"acme-inc": true,
		"acme123":  true,
		"a":        true,
		"":         true, // empty loop body — blank is rejected separately by validateOnboardTenantInput
		"Acme-Inc": false,
		"acme_inc": false,
		"acme inc": false,
		"acme.inc": false,
		"acme/inc": false,
		"ACMEINC":  false,
	}
	for slug, want := range cases {
		if got := isURLSafeSlug(slug); got != want {
			t.Errorf("isURLSafeSlug(%q) = %v, want %v", slug, got, want)
		}
	}
}

func TestAdminEmailOrEmpty(t *testing.T) {
	if got := adminEmailOrEmpty("admin@acme.test"); got != "admin@acme.test" {
		t.Errorf("adminEmailOrEmpty(email) = %q, want %q", got, "admin@acme.test")
	}
	if got := adminEmailOrEmpty("not-an-email"); got != "" {
		t.Errorf("adminEmailOrEmpty(non-email) = %q, want empty", got)
	}
}
