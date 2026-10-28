package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "energy-shares-backend-repo",
	Short: "EnergyShares 공통모듈",
	Long:  `EnergyShares 공통모듈 관련 도구들을 실행합니다.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// 설정 초기화 로직
	viper.AutomaticEnv()
}
