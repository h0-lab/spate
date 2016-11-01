// Copyright © 2016 Matthias Neugebauer <mtneug@mailbox.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/mtneug/spate/api"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spate",
	Short: "Horizontal service autoscaler for Docker Swarm mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()

		addr, err := flags.GetString("listen-address")
		if err != nil {
			return err
		}

		api.Run(addr)

		return nil
	},
}

func init() {
	rootCmd.Flags().String("listen-address", ":8080", "Interface to bind to")

	rootCmd.AddCommand(
		infoCmd,
		versionCmd,
	)
}

// Execute invoces the top-level command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
