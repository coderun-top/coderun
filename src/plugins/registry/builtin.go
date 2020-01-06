package registry

import (
	"github.com/coderun-top/coderun/src/model"
)

type builtin struct {
	store model.RegistryStore
}

// New returns a new local registry service.
func New(store model.RegistryStore) model.RegistryService {
	return &builtin{store}
}

func (b *builtin) RegistryFind(projectName, name string) (*model.Registry, error) {
	return b.store.RegistryFind(projectName, name)
}

func (b *builtin) RegistryList(projectName string) ([]*model.Registry, error) {
	return b.store.RegistryList(projectName)
}

func (b *builtin) RegistryCreate(in *model.Registry) error {
	return b.store.RegistryCreate(in)
}

func (b *builtin) RegistryUpdate(in *model.Registry) error {
	return b.store.RegistryUpdate(in)
}

func (b *builtin) RegistryDelete(projectName, name string) error {
	// registry, err := b.RegistryFind(projectName, name)
	// if err != nil {
	// 	return err
	// }
	return b.store.RegistryDelete(projectName, name)
}
