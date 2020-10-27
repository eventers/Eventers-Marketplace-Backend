package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

type Vault struct {
	UserPath string
	TempPath string
	*api.Client
}

func New(token, unsealKey, address, userPath, tempPath string) (*Vault, error) {
	config := &api.Config{
		Address: address,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("new: error initializing vault: %w", err)
	}

	client.SetToken(token)

	s := client.Sys()
	status, err := s.SealStatus()
	if err != nil {
		return nil, fmt.Errorf("new: error getting seal status: %w", err)
	}

	if !status.Sealed {
		unsealResponse, err := s.Unseal(unsealKey)
		if err != nil {
			return nil, fmt.Errorf("new: error getting unseal response: %w", err)
		}
		if unsealResponse.Sealed {
			return nil, fmt.Errorf("new: vault unseal unsuccesfull")
		}
	}

	err = createIfNotExists(client, userPath)
	if err != nil {
		return nil, fmt.Errorf("new: unable to mount user path: %w", err)
	}

	err = createIfNotExists(client, tempPath)
	if err != nil {
		return nil, fmt.Errorf("new: unable to mount temp path: %w", err)
	}

	return &Vault{UserPath: userPath, TempPath: tempPath, Client: client}, nil
}

func createIfNotExists(client *api.Client, path string) error {
	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return fmt.Errorf("createIfNotExists: unable to list mounts: %w", err)
	}

	if _, ok := mounts[path+"/"]; !ok {
		err = client.Sys().Mount(path, &api.MountInput{Type: "kv"})
		if err != nil {
			return fmt.Errorf("createIfNotExists: unable to create path: %w", err)
		}
	}

	return nil
}
