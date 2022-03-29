package installpackage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/sirupsen/logrus"
)

func (inst *installer) InstallProxy(ctx context.Context) error {
	pkg := &config.Package{
		Name:    proxyName,
		Version: "v0.2.1", // renovate: depName=aquaproj/aqua-proxy
	}
	logE := inst.logE().WithFields(logrus.Fields{
		"package_name":    pkg.Name,
		"package_version": pkg.Version,
		"registry":        pkg.Registry,
	})

	logE.Debug("install the proxy")
	pkgInfo := &config.PackageInfo{
		Type:      "github_release",
		RepoOwner: "aquaproj",
		RepoName:  proxyName,
		Asset:     template.NewTemplate(`aqua-proxy_{{.OS}}_{{.Arch}}.tar.gz`),
		Files: []*config.File{
			{
				Name: proxyName,
			},
		},
	}
	assetName, err := pkgInfo.RenderAsset(pkg)
	if err != nil {
		return err //nolint:wrapcheck
	}

	pkgPath, err := pkgInfo.GetPkgPath(inst.rootDir, pkg)
	if err != nil {
		return err //nolint:wrapcheck
	}
	logE.Debug("check if aqua-proxy is already installed")
	finfo, err := os.Stat(pkgPath)
	if err != nil {
		// file doesn't exist
		if err := inst.downloadWithRetry(ctx, pkg, pkgInfo, pkgPath, assetName); err != nil {
			return err
		}
	} else {
		if !finfo.IsDir() {
			return fmt.Errorf("%s isn't a directory", pkgPath)
		}
	}

	// create a symbolic link
	a, err := filepath.Rel(filepath.Join(inst.rootDir, "bin"), filepath.Join(pkgPath, proxyName))
	if err != nil {
		return fmt.Errorf("get a relative path: %w", err)
	}

	return inst.createLink(filepath.Join(inst.rootDir, "bin", proxyName), a)
}