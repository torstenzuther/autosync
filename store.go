package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

type store interface {
	onWrite(file string, alias string) error
	onRename(file string, alias string) error
	onRemove(file string, alias string) error
	push() error
}

type inMemoryStore struct {
	config *GitConfig
	repo   *git.Repository
	fs     billy.Filesystem
}

func (s *inMemoryStore) push() error {
	return s.repo.Push(&git.PushOptions{
		Auth: getAuthFromConfig(s.config),
	})
}

func getAuthFromConfig(config *GitConfig) transport.AuthMethod {
	if config != nil && config.Auth.UserName != "" && config.Auth.Password != "" {
		return &http.BasicAuth{
			Username: config.Auth.UserName,
			Password: config.Auth.Password,
		}
	}
	return nil
}

func (s *inMemoryStore) commit() error {
	worktree, err := s.repo.Worktree()
	if err != nil {
		return err
	}
	status, err := worktree.Status()
	if err != nil {
		return nil
	}
	if status.IsClean() {
		return nil
	}
	if hash, err := worktree.Commit("autosync", &git.CommitOptions{}); err != nil {
		return err
	} else {
		fmt.Printf("Committed %v\n", hash)
	}
	return nil
}

func (s *inMemoryStore) onWrite(file string, alias string) error {
	fileNameWithoutPath := filepath.Base(file)
	worktree, err := s.repo.Worktree()
	if err != nil {
		return err
	}
	p := path.Join(alias, fileNameWithoutPath)

	osFile, err := os.Open(file)
	defer osFile.Close()
	if err != nil {
		return err
	}
	if f, err := s.fs.Create(p); err != nil {

		return err
	} else {
		var buf = make([]byte, 1000)
		for {
			n, e := osFile.Read(buf)
			if e != nil && e != io.EOF {
				return e
			}
			n2, e2 := f.Write(buf[:n])
			if e2 != nil {
				return e2
			}
			if n2 != n {
				return errors.New("read != write")
			}
			if n == len(buf) {
				continue
			}
			break
		}
	}
	if _, err = worktree.Add(p); err != nil {
		return err
	}
	if err := s.commit(); err != nil {
		return err
	}
	return nil
}

func (s *inMemoryStore) onRename(file string, alias string) error {
	if _, err := os.Stat(file); err == nil || !os.IsNotExist(err) {
		return s.onWrite(file, alias)
	}
	return s.onRemove(file, alias)
}

func (s *inMemoryStore) onRemove(file string, alias string) error {
	fileNameWithoutPath := filepath.Base(file)
	worktree, err := s.repo.Worktree()
	if err != nil {
		return err
	}
	p := path.Join(alias, fileNameWithoutPath)

	fmt.Printf("Remove %v\n", p)
	if _, err = worktree.Remove(p); err != nil {
		return err
	}
	if err := s.commit(); err != nil {
		return err
	}
	return nil
}

func newInMemoryStore(cfg *Config) (store, error) {
	fs := memfs.New()
	repo, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:  cfg.GitRepo.Url,
		Auth: getAuthFromConfig(&cfg.GitRepo),
	})

	for _, pathMapping := range cfg.PathMappings {
		if err := fs.MkdirAll(pathMapping.GitPath, os.ModeDir); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return &inMemoryStore{
		repo:   repo,
		fs:     fs,
		config: &cfg.GitRepo,
	}, nil
}
