package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	rootPath string
}

func NewStore(rootPath string) *Store {
	return &Store{rootPath: rootPath}
}

func (s *Store) Save(_ context.Context, storageKey string, content []byte) error {
	target, err := s.targetPath(storageKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("mkdir attachment path: %w", err)
	}
	if err := os.WriteFile(target, content, 0o644); err != nil {
		return fmt.Errorf("write attachment content: %w", err)
	}
	return nil
}

func (s *Store) Open(_ context.Context, storageKey string) (io.ReadCloser, error) {
	target, err := s.targetPath(storageKey)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(target)
	if err != nil {
		return nil, fmt.Errorf("open attachment content: %w", err)
	}
	return file, nil
}

func (s *Store) Delete(_ context.Context, storageKey string) error {
	target, err := s.targetPath(storageKey)
	if err != nil {
		return err
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete attachment content: %w", err)
	}
	return nil
}

func (s *Store) targetPath(storageKey string) (string, error) {
	if strings.ContainsRune(storageKey, 0) {
		return "", fmt.Errorf("local store: key contains null byte")
	}

	root, err := filepath.Abs(filepath.Clean(s.rootPath))
	if err != nil {
		return "", fmt.Errorf("local store: resolve root path: %w", err)
	}

	target := filepath.Clean(filepath.Join(root, filepath.FromSlash(storageKey)))
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", fmt.Errorf("local store: resolve key path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("local store: key %q escapes root", storageKey)
	}

	return target, nil
}
