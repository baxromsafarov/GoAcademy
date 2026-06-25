// Package storage abstracts blob storage for user uploads (avatars, ...).
// The local implementation writes to a directory and serves files over /static;
// a future S3 implementation can satisfy the same interface (decision D-009).
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Storage stores and removes objects addressed by a forward-slash key
// (e.g. "avatars/<id>.png") and returns a public URL for a saved object.
type Storage interface {
	Save(ctx context.Context, key string, data io.Reader) (url string, err error)
	Delete(ctx context.Context, key string) error
}

// LocalStorage writes objects under baseDir and builds public URLs from
// publicBaseURL (the path served by the static file handler).
type LocalStorage struct {
	baseDir       string
	publicBaseURL string
}

// NewLocalStorage creates the base directory if needed and returns a LocalStorage.
func NewLocalStorage(baseDir, publicBaseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &LocalStorage{baseDir: baseDir, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}, nil
}

func (s *LocalStorage) Save(_ context.Context, key string, data io.Reader) (string, error) {
	dest, err := s.resolve(key)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("create object dir: %w", err)
	}
	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create object: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		return "", fmt.Errorf("write object: %w", err)
	}
	return s.publicBaseURL + "/" + key, nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	dest, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// resolve maps a key to an absolute path under baseDir, rejecting traversal.
func (s *LocalStorage) resolve(key string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(key))
	if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", fmt.Errorf("invalid storage key: %q", key)
	}
	return filepath.Join(s.baseDir, clean), nil
}
