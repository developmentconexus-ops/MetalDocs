package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strings"

	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/tenant"
)

type seedConfig struct {
	UserID      string
	Username    string
	Email       string
	DisplayName string
	Password    string
}

func main() {
	os.Exit(run())
}

// run performs the e2e seed and returns the process exit code. Body of the
// former main(): extracted so os.Exit is called exactly once, after run
// returns, letting `defer deps.Cleanup()` (and any later defers) run on
// every early-return path instead of being skipped by an in-place os.Exit
// (gocritic exitAfterDefer).
func run() int {
	ctx := context.Background()

	repoMode, err := config.RepositoryMode()
	if err != nil {
		log.Printf("invalid repository mode: %v", err)
		return 1
	}
	if repoMode != config.RepositoryPostgres {
		log.Printf("metaldocs-e2e-seed requires postgres repository mode")
		return 1
	}
	attachmentsCfg, err := config.LoadAttachmentsConfig()
	if err != nil {
		log.Printf("invalid attachments config: %v", err)
		return 1
	}
	authCfg, err := authn.LoadRuntimeConfig()
	if err != nil {
		log.Printf("invalid auth config: %v", err)
		return 1
	}

	deps, err := bootstrap.BuildAPIDependencies(ctx, repoMode, attachmentsCfg)
	if err != nil {
		log.Printf("build api dependencies: %v", err)
		return 1
	}
	defer deps.Cleanup()

	authService, err := authapp.NewService(deps.AuthRepo, deps.RoleProvider, deps.RoleAdminRepo, iampg.NewLoginContextRepository(deps.SQLDB), authCfg)
	if err != nil {
		log.Printf("new auth service: %v", err)
		return 1
	}
	iamAdmin := iamapp.NewAdminService(deps.RoleAdminRepo, nil, nil, nil)
	seed := loadSeedConfig()

	exists, err := userExists(ctx, authService, seed.UserID)
	if err != nil {
		log.Printf("check existing user: %v", err)
		return 1
	}

	if !exists {
		if err := authService.CreateUser(ctx, seed.UserID, seed.Username, seed.Email, seed.DisplayName, seed.Password, tenant.DevTenantID, []iamdomain.Role{iamdomain.RoleSystemAdmin}, "e2e-seed"); err != nil {
			log.Printf("create e2e user: %v", err)
			return 1
		}
	} else {
		active := true
		mustChangePassword := false
		email := seed.Email
		displayName := seed.DisplayName
		if err := authService.UpdateUser(ctx, authdomain.UpdateUserParams{
			UserID:             seed.UserID,
			DisplayName:        &displayName,
			Email:              &email,
			IsActive:           &active,
			MustChangePassword: &mustChangePassword,
		}, seed.Password); err != nil {
			log.Printf("reset e2e user: %v", err)
			return 1
		}
	}

	if err := iamAdmin.UpsertUserAndAssignRole(ctx, seed.UserID, seed.DisplayName, tenant.DevTenantID, iamdomain.RoleSystemAdmin, "e2e-seed", "e2e-seed"); err != nil {
		log.Printf("ensure admin role: %v", err)
		return 1
	}

	slog.Info("e2e seed ready", "user_id", seed.UserID, "username", seed.Username)
	return 0
}

func loadSeedConfig() seedConfig {
	return seedConfig{
		UserID:      readEnv("METALDOCS_E2E_ADMIN_USER_ID", "e2e-admin"),
		Username:    readEnv("METALDOCS_E2E_ADMIN_USERNAME", "e2e.admin"),
		Email:       readEnv("METALDOCS_E2E_ADMIN_EMAIL", "e2e.admin@local.test"),
		DisplayName: readEnv("METALDOCS_E2E_ADMIN_DISPLAY_NAME", "E2E Admin"),
		Password:    requireEnv("METALDOCS_E2E_ADMIN_PASSWORD", "E2eAdmin123!"),
	}
}

func userExists(ctx context.Context, service *authapp.Service, userID string) (bool, error) {
	items, err := service.ListUsers(ctx, tenant.DevTenantID)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if strings.TrimSpace(item.UserID) == strings.TrimSpace(userID) {
			return true, nil
		}
	}
	return false, nil
}

func readEnv(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func requireEnv(name, fallback string) string {
	value := readEnv(name, fallback)
	if strings.TrimSpace(value) == "" {
		log.Fatalf("%s is required", name)
	}
	return value
}
