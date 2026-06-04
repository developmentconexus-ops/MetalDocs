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

// capChecker is a struct fake for application.CapabilityChecker.
type capChecker struct {
	deny  error
	admin bool
}

func (f capChecker) CanDo(_ context.Context, _, _ string, _ iamdomain.Capability) error {
	return f.deny
}

func (f capChecker) IsSystemAdmin(_ context.Context, _, _ string) (bool, error) {
	return f.admin, nil
}

func TestCreateDocument_DeniesWhenCapabilityChecker_Denies(t *testing.T) {
	denying := capChecker{deny: iamapp.ErrCapabilityDenied}

	cd := &controlleddocumentsdomain.ControlledDocument{
		ID:              "cd_1",
		ProfileCode:     "PROC",
		ProcessAreaCode: "AREA-01",
		Status:          controlleddocumentsdomain.CDStatusActive,
	}

	svc := application.NewService(
		&fakeRepo{},
		&fakePresigner{},
		fakeTplReader{},
		fakeFormVal{valid: true},
		&noopAudit{},
		&fakeControlledDocumentReader{cd: cd},
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
