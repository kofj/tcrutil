package tcrutil

import (
	"errors"

	"github.com/goharbor/harbor/src/lib/log"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
)

func (t *tcrutil) CreatePrivateNamespace(namespace string) (err error) {
	if t.tcrClient == nil {
		err = errors.New("[tcr.createPrivateNamespace] nil tcr client")
		return
	}

	// 1. if exist skip
	log.Debugf("[tcr.PrepareForPush.createPrivateNamespace] namespace=%s", namespace)
	var exist bool
	exist, err = t.IsNamespaceExist(namespace)
	if err != nil {
		return
	}
	if exist {
		log.Warningf("[tcr.PrepareForPush.createPrivateNamespace.skip_exist] namespace=%s", namespace)
		return
	}

	// !!! 2. WARNING: for safty, auto create namespace is private.
	var req = tcr.NewCreateNamespaceRequest()
	req.NamespaceName = &namespace
	req.RegistryId = t.registryID
	var isPublic = false
	req.IsPublic = &isPublic
	tcr.NewCreateNamespaceResponse()
	_, err = t.tcrClient.CreateNamespace(req)
	if err != nil {
		log.Debugf("[tcr.PrepareForPush.createPrivateNamespace] error=%v", err)
		return
	}
	return
}

func (a *tcrutil) CreateRepository(namespace, repository string) (err error) {
	if a.tcrClient == nil {
		err = errors.New("[tcr.createRepository] nil tcr client")
		return
	}

	// 1. if exist skip
	log.Debugf("[tcr.PrepareForPush.createRepository] namespace=%s, repository=%s", namespace, repository)
	var repoReq = tcr.NewDescribeRepositoriesRequest()
	repoReq.RegistryId = a.registryID
	repoReq.NamespaceName = &namespace
	repoReq.RepositoryName = &repository
	var repoResp = tcr.NewDescribeRepositoriesResponse()
	repoResp, err = a.tcrClient.DescribeRepositories(repoReq)
	if err != nil {
		return
	}
	if int(*repoResp.Response.TotalCount) > 0 {
		log.Warningf("[tcr.PrepareForPush.createRepository.skip_exist] namespace=%s, repository=%s", namespace, repository)
		return
	}

	// 2. create
	var req = tcr.NewCreateRepositoryRequest()
	req.NamespaceName = &namespace
	req.RepositoryName = &repository
	req.RegistryId = a.registryID
	var resp = tcr.NewCreateRepositoryResponse()
	resp, err = a.tcrClient.CreateRepository(req)
	if err != nil {
		log.Debugf("[tcr.PrepareForPush.createRepository] error=%v", err)
		return
	}
	log.Debugf("[tcr.PrepareForPush.createRepository] resp=%#v", *resp)

	return
}

func (a *tcrutil) ListNamespaces() (namespaces []string, err error) {
	if a.tcrClient == nil {
		err = errors.New("[tcr.listNamespaces] nil tcr client")
		return
	}

	// list namespaces
	var req = tcr.NewDescribeNamespacesRequest()
	req.RegistryId = a.registryID
	req.Limit = a.pageSize
	var resp = tcr.NewDescribeNamespacesResponse()

	var page int64
	for {
		req.Offset = &page
		resp, err = a.tcrClient.DescribeNamespaces(req)
		if err != nil {
			log.Debugf("[tcr.DescribeNamespaces] registryID=%s, error=%v", *a.registryID, err)
			return
		}

		for _, ns := range resp.Response.NamespaceList {
			namespaces = append(namespaces, *ns.Name)
		}

		if len(namespaces) >= int(*resp.Response.TotalCount) {
			break
		}
		page++
	}

	log.Debugf("[tcr.FetchArtifacts.listNamespaces] registryID=%s, namespaces[%d]=%s", *a.registryID, len(namespaces), namespaces)
	return
}

func (a *tcrutil) IsNamespaceExist(namespace string) (exist bool, err error) {
	if a.tcrClient == nil {
		err = errors.New("[tcr.isNamespaceExist] nil tcr client")
		return
	}

	var req = tcr.NewDescribeNamespacesRequest()
	req.NamespaceName = &namespace
	req.RegistryId = a.registryID
	var resp = tcr.NewDescribeNamespacesResponse()
	resp, err = a.tcrClient.DescribeNamespaces(req)
	if err != nil {
		return
	}

	log.Warningf("[tcr.PrepareForPush.isNamespaceExist] namespace=%s, total=%d", namespace, *resp.Response.TotalCount)
	if int(*resp.Response.TotalCount) != 1 {
		return
	}
	exist = true
	return
}

func (a *tcrutil) ListReposByNamespace(namespace string) (repos []string, err error) {
	if a.tcrClient == nil {
		err = errors.New("[tcr.listReposByNamespace] nil tcr client")
		return
	}

	var req = tcr.NewDescribeRepositoriesRequest()
	req.RegistryId = a.registryID
	req.NamespaceName = &namespace
	req.Limit = a.pageSize
	var resp = tcr.NewDescribeRepositoriesResponse()

	var page int64
	for {
		req.Offset = &page
		resp, err = a.tcrClient.DescribeRepositories(req)
		if err != nil {
			log.Debugf("[tcr.DescribeRepositories] registryID=%s, namespace=%s, error=%v", *a.registryID, namespace, err)
			return
		}

		for _, repo := range resp.Response.RepositoryList {
			repos = append(repos, *repo.Name)
		}

		if len(repos) == int(*resp.Response.TotalCount) {
			break
		}
		page++
	}

	log.Debugf("[tcr.listReposByNamespace] registryID=%s, namespace=%s, repos[%d]=%v",
		*a.registryID, namespace, len(repos), repos)
	return
}

func (a *tcrutil) GetImages(namespace, repo string) (images []string, err error) {
	if a.tcrClient == nil {
		err = errors.New("[tcr.getImages] nil tcr client")
		return
	}

	var req = tcr.NewDescribeImagesRequest()
	req.RegistryId = a.registryID
	// ! if repoName include namespace, keep namespace empty
	req.NamespaceName = &namespace
	req.RepositoryName = &repo
	req.Limit = a.pageSize
	var resp = tcr.NewDescribeImagesResponse()

	var page int64
	for {
		req.Offset = &page
		resp, err = a.tcrClient.DescribeImages(req)
		if err != nil {
			log.Debugf("[tcr.getImages.DescribeImages] registryID=%s, namespace=%s, repo=%s, error=%v", *a.registryID, namespace, repo, err)
			return
		}

		for _, image := range resp.Response.ImageInfoList {
			images = append(images, *image.ImageVersion)
		}

		if len(images) == int(*resp.Response.TotalCount) {
			break
		}
		page++
	}

	log.Debugf("[tcr.getImages] registryID=%s, namespace=%s, repo=%s, tags[%d]=%v\n", *a.registryID, namespace, repo, len(images), images)
	return
}
