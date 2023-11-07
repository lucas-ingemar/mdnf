package cmd

import (
	"fmt"

	"github.com/lucas-ingemar/packtrak/internal/machinery"
	"github.com/lucas-ingemar/packtrak/internal/packagemanagers"
	"github.com/lucas-ingemar/packtrak/internal/shared"
	"github.com/lucas-ingemar/packtrak/internal/state"
	"github.com/spf13/cobra"
)

func initList() {
	for _, pm := range packagemanagers.PackageManagers {
		PmCmds[pm.Name()].AddCommand(&cobra.Command{
			Use:   "list",
			Short: fmt.Sprintf("List status of %s packages", pm.Name()),
			Args:  cobra.NoArgs,
			Run:   generateListCmd([]shared.PackageManager{pm}),
		})
	}

	var listGlobalCmd = &cobra.Command{
		Use:   "list",
		Short: "List status of dnf packages",
		Args:  cobra.NoArgs,
		Run:   generateListCmd(packagemanagers.PackageManagers),
	}
	rootCmd.AddCommand(listGlobalCmd)
}

func generateListCmd(pms []shared.PackageManager) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if !shared.MustDoSudo(cmd.Context(), pms, shared.CommandList) {
			panic("sudo access not granted")
		}

		tx := state.Begin()

		depStatus, pkgStatus, err := machinery.ListStatus(cmd.Context(), tx, pms)
		if err != nil {
			panic(err)
		}

		res := tx.Commit()
		if res.Error != nil {
			panic(res.Error)
		}

		machinery.PrintPackageList(depStatus, pkgStatus)
	}
}
