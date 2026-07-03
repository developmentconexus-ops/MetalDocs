package domain

import "errors"

var (
	// ErrUserNotFound is returned when no iam_users row exists for the requested
	// (tenant, user) pair.
	ErrUserNotFound = errors.New("iam user not found")
	// ErrUserInactive is returned when the resolved user exists but is
	// deactivated — callers must treat this the same as "not found" for
	// authorization purposes.
	ErrUserInactive = errors.New("iam user inactive")
	// ErrNoRolesAssigned is returned when a user has no role assignments in the
	// tenant (RoleProvider.RolesByUserID and related lookups).
	ErrNoRolesAssigned = errors.New("iam no roles assigned")
	// ErrInvalidRole is returned by ParseRole when the input does not match one
	// of the eight canonical roles.
	ErrInvalidRole = errors.New("iam invalid role")
)
