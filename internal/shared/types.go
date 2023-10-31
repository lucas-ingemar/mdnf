package shared

import (
	"fmt"
	"time"
)

type Packages map[string]PmPackages

func (p Packages) Register(packageManagerName string) error {
	_, exists := p[packageManagerName]
	if exists {
		return fmt.Errorf("%s already exists", packageManagerName)
	}
	p[packageManagerName] = PmPackages{}
	return nil
}

type PmPackages struct {
	Global PackagesGlobal `yaml:"global"`
}

type PackagesGlobal struct {
	Packages []string `yaml:"packages"`
}

type State struct {
	Timestamp time.Time `yaml:"timestamp"`
	Packages  Packages  `yaml:"packages"`
}
