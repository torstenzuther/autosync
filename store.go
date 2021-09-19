package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

type store interface {
	onCreateEvent(file string) error
	onRenameEvent(file string) error
	onWriteEvent(file string) error
	onRemoveEvent(file string) error
	commit() error
}

type inMemoryStore struct {
	repo *git.Repository
}

func (s *inMemoryStore) commit() error {
	worktree, err := s.repo.Worktree()
	if err != nil {
		return err
	}
	if _, err = worktree.Commit("autosync", &git.CommitOptions{}); err != nil {
		return err
	}
	return nil
}

func (s *inMemoryStore) onCreateEvent(file string) error {
	worktree, err := s.repo.Worktree()
	if err != nil {
		return err
	}
	if _, err = worktree.Add(file); err != nil {
		return err
	}
	return nil
}

func (s *inMemoryStore) onRenameEvent(file string) error {
	//worktree, err := s.repo.Worktree()
	//if err != nil {
	//	return err
	//}
	//if _, err = worktree.Re(file); err != nil {
	//	return err
	//}
	return nil
}

func (s *inMemoryStore) onWriteEvent(file string) error {
	worktree, err := s.repo.Worktree()
	if err != nil {
		return err
	}
	if _, err = worktree.Add(file); err != nil {
		return err
	}
	return nil
}

func (s *inMemoryStore) onRemoveEvent(file string) error {
	worktree, err := s.repo.Worktree()
	if err != nil {
		return err
	}
	if _, err = worktree.Remove(file); err != nil {
		return err
	}
	return nil
}

func newInMemoryStore() (store, error) {
	repo, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		return nil, err
	}
	return &inMemoryStore{repo: repo}, nil
}
