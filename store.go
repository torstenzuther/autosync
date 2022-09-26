package main

import (
	"errors"
	"io"
	"log"
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
	commit() error
}

type inMemoryStore struct {
	config *GitConfig
	repo   *git.Repository
	fs     billy.Filesystem
}

const commitMessage = "autosync"
const fileReadBufferSize = 10000

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
		return err
	}
	if status.IsClean() {
		return nil
	}
	hash, err := worktree.Commit(commitMessage, &git.CommitOptions{})
	if err != nil {
		return err
	}
	log.Printf("Commit %v\n", hash)
	return nil
}

func (s *inMemoryStore) onWrite(file string, alias string) error {
	if err := s.writeWithoutCommit(file, alias); err != nil {
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
	if _, err = worktree.Remove(p); err != nil {
		return err
	}
	return nil
}

func (s *inMemoryStore) writeWithoutCommit(file string, alias string) (err error) {
	fileNameWithoutPath := filepath.Base(file)
	var worktree *git.Worktree
	worktree, err = s.repo.Worktree()
	if err != nil {
		return
	}
	p := path.Join(alias, fileNameWithoutPath)

	var osFile *os.File
	osFile, err = os.Open(file)
	defer func(f *os.File) {
		err = f.Close()
	}(osFile)
	if err != nil {
		return
	}
	var f billy.File
	if f, err = s.fs.Create(p); err != nil {
		return
	}
	defer func(f billy.File) {
		err = f.Close()
	}(f)

	var readBuf = make([]byte, fileReadBufferSize)
	var bytesRead int
	for bytesRead < len(readBuf) && err != io.EOF {
		bytesRead, err = osFile.Read(readBuf)
		if err != nil && err != io.EOF {
			return
		}
		bytesWritten, writeErr := f.Write(readBuf[:bytesRead])
		if writeErr != nil {
			return writeErr
		}
		if bytesWritten != bytesRead {
			return errors.New("read != write")
		}
	}
	if _, err = worktree.Add(p); err != nil {
		return
	}
	return nil
}

func newInMemoryStore(cfg *Config) (store, error) {
	filesystem := memfs.New()
	repo, err := git.Clone(memory.NewStorage(), filesystem, &git.CloneOptions{
		URL:  cfg.GitRepo.Url,
		Auth: getAuthFromConfig(&cfg.GitRepo),
	})
	store := &inMemoryStore{
		repo:   repo,
		fs:     filesystem,
		config: &cfg.GitRepo,
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	if err := worktree.RemoveGlob("*"); err != nil {
		return nil, err
	}
	for _, pathMapping := range cfg.PathMappings {
		if err := filesystem.MkdirAll(pathMapping.GitPath, os.ModeDir); err != nil {
			return nil, err
		}
		patternAbs, err := filepath.Abs(pathMapping.Pattern)
		if err != nil {
			log.Printf("%v\n", err)
		}
		actualDir := filepath.Dir(patternAbs)
		dirInfos, err := os.ReadDir(actualDir)
		if err != nil {
			return nil, err
		}
		for _, dirInfo := range dirInfos {
			if dirInfo.IsDir() {
				continue
			}
			abs := filepath.Join(actualDir, dirInfo.Name())
			ok, err := filepath.Match(patternAbs, abs)
			if err != nil {
				return nil, err
			}
			if ok {
				if err := store.writeWithoutCommit(abs, pathMapping.GitPath); err != nil {
					return nil, err
				}
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return store, nil
}
