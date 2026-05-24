package local

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreSaveRejectsEscapingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	root := t.TempDir()
	outside, key := escapingSibling(root, "outside.txt")
	store := NewStore(root)

	err := store.Save(ctx, key, []byte("escape"))
	if err == nil {
		t.Fatal("expected escaping key to fail")
	}
	if _, statErr := os.Stat(outside); !os.IsNotExist(statErr) {
		t.Fatalf("outside file was touched: %v", statErr)
	}
}

func TestStoreOpenRejectsEscapingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	root := t.TempDir()
	outside, key := escapingSibling(root, "outside.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	store := NewStore(root)

	file, err := store.Open(ctx, key)
	if err == nil {
		_ = file.Close()
		t.Fatal("expected escaping key to fail")
	}
}

func TestStoreDeleteRejectsEscapingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	root := t.TempDir()
	outside, key := escapingSibling(root, "outside.txt")
	if err := os.WriteFile(outside, []byte("keep"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	store := NewStore(root)

	err := store.Delete(ctx, key)
	if err == nil {
		t.Fatal("expected escaping key to fail")
	}
	if _, statErr := os.Stat(outside); statErr != nil {
		t.Fatalf("outside file was removed or changed: %v", statErr)
	}
}

func TestStoreRejectsNullByteKey(t *testing.T) {
	t.Parallel()

	store := NewStore(t.TempDir())

	_, err := store.targetPath("safe" + string(rune(0)) + "key")
	if err == nil {
		t.Fatal("expected null byte key to fail")
	}
	if !strings.Contains(err.Error(), "null byte") {
		t.Fatalf("expected null byte error, got %v", err)
	}
}

func TestStoreAllowsNestedKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	root := t.TempDir()
	store := NewStore(root)

	if err := store.Save(ctx, "tenant/documents/file.txt", []byte("ok")); err != nil {
		t.Fatalf("save nested key: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "tenant", "documents", "file.txt"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(got) != "ok" {
		t.Fatalf("saved content = %q, want ok", got)
	}
}

func escapingSibling(root, name string) (string, string) {
	siblingName := filepath.Base(root) + "-" + name
	return filepath.Join(filepath.Dir(root), siblingName), "../" + siblingName
}
