package tcrutil

import (
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
)

// Tcrutil TCR util interface
type Tcrutil interface {
	CreatePrivateNamespace(namespace string) (err error)
	CreateRepository(namespace, repository string) (err error)
	GetImages(namespace, repo string) (images []string, err error)
	IsNamespaceExist(namespace string) (exist bool, err error)
	ListReposByNamespace(namespace string) (repos []string, err error)
	ListNamespaces() (namespaces []string, err error)
}

type tcrutil struct {
	tcrClient  *tcr.Client
	registryID *string
	pageSize   *int64
}

var _ Tcrutil = &tcrutil{}

// New TCR client
func New(tcr *tcr.Client, registryID *string, pageSize *int64) Tcrutil {
	return &tcrutil{
		tcrClient:  tcr,
		pageSize:   pageSize,
		registryID: registryID,
	}
}
