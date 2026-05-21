package application_test

import (
	"context"
	"errors"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// capCheckerFunc is a function adapter for application.CapabilityChecker.
type capCheckerFunc func(ctx context.Context, userID, tenantID string, cap iamdomain.Capability) error

func (f capCheckerFunc) CanDo(ctx context.Context, u, t string, c iamdomain.Capability) error {
	return f(ctx, u, t, c)
}

func TestCreateDocument_DeniesWhenCapabilityChecker_Denies(t *testing.T) {
	denying := capCheckerFunc(func(ctx context.Context, userID, tenantID string, cap iamdomain.Capability) error {
		return iamapp.ErrCapabilityDenied
	})

	cd := &controlleddocumentsdomain.ControlledDocument{
		ID:              "cd_1",
		ProfileCode:     "PROC",
		ProcessAreaCode: "AREA-01",
		Status:          controlleddocumentsdomain.CDStatusActive,
	}

	svc := application.NewService(
		&fakeRepo{},
		fakeDocgen{},
		&fakePresigner{},
		fakeTplReader{},
		fakeFormVal{valid: true},
		&noopAudit{},
		&fakeRegistryReader{cd: cd},
		denying,
		&fakeProfileDefaultTemplateReader{id: strptr("tpl_ver_1"), status: strptr("published")},
	)

	_, err := svc.CreateDocument(context.Background(), application.CreateDocumentInput{
		TenantID:             "t1",
		ActorUserID:          "u1",
		ControlledDocumentID: "cd_1",
	})

	if !errors.Is(err, iamapp.ErrCapabilityDenied) {
		t.Fatalf("got %v, want ErrCapabilityDenied", err)
	}
}
