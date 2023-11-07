package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/config"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync DNF to match mDNF",
	Args:  cobra.NoArgs,
	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmdSync(cmd.Context(), packagemanagers.PackageManagers)
		if err != nil {
			panic(err)
		}
	},
}

func cmdSync(ctx context.Context, pms []shared.PackageManager) (err error) {
	if !shared.MustDoSudo(ctx, pms, shared.CommandSync) {
		return errors.New("sudo access not granted")
	}

	tx := state.Begin()
	defer tx.Rollback()

	totUpdatedPkgs := []shared.Package{}
	totUpdatedDeps := []shared.Dependency{}

	pkgsState := map[string][]shared.Package{}
	depsState := map[string][]shared.Dependency{}

	depStatus, pkgStatus, err := listStatus(ctx, tx, pms)
	if err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	for _, pm := range pms {
		totUpdatedDeps = append(totUpdatedDeps, depStatus[pm.Name()].Missing...)
		totUpdatedDeps = append(totUpdatedDeps, depStatus[pm.Name()].Updated...)
		totUpdatedDeps = append(totUpdatedDeps, depStatus[pm.Name()].Removed...)

		depsState[pm.Name()] = append(depsState[pm.Name()], depStatus[pm.Name()].Synced...)
		depsState[pm.Name()] = append(depsState[pm.Name()], depStatus[pm.Name()].Updated...)
		depsState[pm.Name()] = append(depsState[pm.Name()], depStatus[pm.Name()].Missing...)

		totUpdatedPkgs = append(totUpdatedPkgs, pkgStatus[pm.Name()].Missing...)
		totUpdatedPkgs = append(totUpdatedPkgs, pkgStatus[pm.Name()].Updated...)
		totUpdatedPkgs = append(totUpdatedPkgs, pkgStatus[pm.Name()].Removed...)

		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Synced...)
		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Updated...)
		pkgsState[pm.Name()] = append(pkgsState[pm.Name()], pkgStatus[pm.Name()].Missing...)
	}

	printPackageList(depStatus, pkgStatus)

	if len(totUpdatedDeps) == 0 && len(totUpdatedPkgs) == 0 {
		tx := state.Begin()
		defer tx.Rollback()
		for _, pm := range pms {
			err := state.UpdatePackageState(tx, pm.Name(), pkgsState[pm.Name()])
			if err != nil {
				return err
			}

			err = state.UpdateDependencyState(tx, pm.Name(), depsState[pm.Name()])
			if err != nil {
				return err
			}
		}
		if err := tx.Commit().Error; err != nil {
			return err
		}
		return state.Rotate(config.StateRotations)
	}

	fmt.Println("")
	result, _ := pterm.InteractiveContinuePrinter{
		DefaultValueIndex: 0,
		DefaultText:       "Unsynced changes found in config. Do you want to sync?",
		TextStyle:         &pterm.ThemeDefault.PrimaryStyle,
		Options:           []string{"y", "n"},
		OptionsStyle:      &pterm.ThemeDefault.SuccessMessageStyle,
		SuffixStyle:       &pterm.ThemeDefault.SecondaryStyle,
		Delimiter:         ": ",
	}.Show()

	if result == "y" {
		for _, pm := range pms {
			tx := state.Begin()
			defer tx.Rollback()

			uw, err := pm.SyncDependencies(ctx, depStatus[pm.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = state.UpdateDependencyState(tx, pm.Name(), depsState[pm.Name()])
			if err != nil {
				return err
			}

			if err := tx.Commit().Error; err != nil {
				return err
			}

			tx = state.Begin()
			defer tx.Rollback()

			uw, err = pm.SyncPackages(ctx, pkgStatus[pm.Name()])
			_ = uw
			if err != nil {
				return err
			}
			err = state.UpdatePackageState(tx, pm.Name(), pkgsState[pm.Name()])
			if err != nil {
				return err
			}

			if err := tx.Commit().Error; err != nil {
				return err
			}
		}
	}

	return state.Rotate(config.StateRotations)
}
