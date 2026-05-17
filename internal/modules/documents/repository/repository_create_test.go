package repository

import (
	"os"
	"strings"
	"testing"
)

func TestCreateDocumentInsertIncludesBridgeSnapshots(t *testing.T) {
	src, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatalf("read repository.go: %v", err)
	}

	body := string(src)
	start := strings.Index(body, "func (r *Repository) CreateDocument")
	if start == -1 {
		t.Fatal("CreateDocument not found")
	}
	end := strings.Index(body[start:], "func (r *Repository) SetRevisionStorageKey")
	if end == -1 {
		t.Fatal("SetRevisionStorageKey not found")
	}
	createDocument := body[start : start+end]

	tests := []struct {
		name string
		want string
	}{
		{name: "profile column", want: "profile_code_snapshot"},
		{name: "process area column", want: "process_area_code_snapshot"},
		{name: "profile arg", want: "d.ProfileCodeSnapshot"},
		{name: "process area arg", want: "d.ProcessAreaCodeSnapshot"},
		{name: "revision number column", want: "revision_number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(createDocument, tt.want) {
				t.Fatalf("CreateDocument missing %s", tt.want)
			}
		})
	}
}

func TestCreateDocumentTx_AssertsEditBeforeDocumentUpdates(t *testing.T) {
	src, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatalf("read repository.go: %v", err)
	}

	body := string(src)
	start := strings.Index(body, "func (r *Repository) CreateDocumentTx")
	if start == -1 {
		t.Fatal("CreateDocumentTx not found")
	}
	end := strings.Index(body[start:], "func (r *Repository) SetRevisionStorageKey")
	if end == -1 {
		t.Fatal("SetRevisionStorageKey not found")
	}
	createDocumentTx := body[start : start+end]

	editRequire := strings.Index(createDocumentTx, "CapDocumentEdit")
	if editRequire == -1 {
		t.Fatal("CreateDocumentTx must assert document.edit before updating documents")
	}
	updateDocuments := strings.Index(createDocumentTx, "UPDATE documents SET current_revision_id")
	if updateDocuments == -1 {
		t.Fatal("CreateDocumentTx documents pointer update not found")
	}
	if editRequire > updateDocuments {
		t.Fatal("CreateDocumentTx asserts document.edit after documents update; tripwire requires it before UPDATE")
	}
}
