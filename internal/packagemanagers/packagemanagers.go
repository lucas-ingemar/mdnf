package packagemanagers

import (
	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/shared"
)

var (
	PackageManagersRegistered = []shared.PackageManager{&Dnf{}, &Go{}}
	PackageManagers           = []shared.PackageManager{}
)

// type PackageManager interface {
// 	Name() string
// 	Icon() string
// 	NeedsSudo() []shared.CommandName

// 	GetPackageNames(ctx context.Context, packagesConfig shared.PmPackages) []string

// 	// FIXME: Update to new format
// 	Add(ctx context.Context, packagesConfig shared.PmPackages, pkgs []string) (packagesConfigUpdated shared.PmPackages, userWarnings []string, err error)
// 	InstallValidArgs(ctx context.Context, toComplete string) ([]string, error)
// 	ListDependencies(ctx context.Context, tx *gorm.DB, packages shared.PmPackages) (depStatus shared.DependenciesStatus, err error)
// 	ListPackages(ctx context.Context, tx *gorm.DB, packages shared.PmPackages) (packageStatus shared.PackageStatus, err error)
// 	// FIXME: Update to new format
// 	Remove(ctx context.Context, packagesConfig shared.PmPackages, pkgs []string) (packagesConfigUpdated shared.PmPackages, userWarnings []string, err error)
// 	SyncDependencies(ctx context.Context, depStatus shared.DependenciesStatus) (userWarnings []string, err error)
// 	SyncPackages(ctx context.Context, packageStatus shared.PackageStatus) (userWarnings []string, err error)
// }

func InitPackageManagers() {
	for _, pm := range PackageManagersRegistered {
		//FIXME: HERE should the enabled/disabled flag be
		PackageManagers = append(PackageManagers, pm)
	}
}

func MustInitPackages() shared.Packages {
	var err error
	config.Packages, err = config.ReadPackagesConfig()
	if err != nil {
		panic(err)
	}

	for _, pm := range PackageManagers {
		_, ok := config.Packages[pm.Name()]
		if !ok {
			config.Packages[pm.Name()] = shared.PmPackages{}
		}
	}
	return config.Packages
}
