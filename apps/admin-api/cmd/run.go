/*
Copyright © 2019 Illuminarean <junyong.park@illuminarean.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/illuminarean/energy-share-core-api/migration"
	"github.com/illuminarean/energy-share-core-api/server"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Energy Shares Admin API 서버를 시작합니다. ",
	Long:  `Energy Shares Admin API 서버를 시작합니다.`,
	Run: func(cmd *cobra.Command, args []string) {

		env, _ := cmd.Flags().GetString("env")

		fmt.Println("Energy Shares Admin API Server Starting.....")
		server.RunServer(env)
	},
}

var swaggerCmd = &cobra.Command{
	Use:   "swagger",
	Short: "Energy Shares Admin API Swagger 문서를 생성합니다.",
	Long:  `Energy Shares Admin API Swagger 문서를 생성합니다.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Energy Shares Admin API Swagger JSON generating.....")

		server.NewBootstrap(
			echo.New(),
			&config.Config{
				CommonConfig: config.CommonConfig{Debug: true},
			},
		).Invoke()
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Energy Shares Database 마이그레이션 작업",
	Long:  `Energy Shares Database 마이그레이션 작업을 시작합니다.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetString("env")

		fmt.Println("Energy Shares Admin API Database migration Starting.....")
		migration.MigrateDatabase(env)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(swaggerCmd)
	rootCmd.AddCommand(migrateCmd)
	runCmd.Flags().StringP("env", "e", viper.GetString("ENV"), "set server config")
	migrateCmd.Flags().StringP("env", "e", viper.GetString("ENV"), "set server config")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
