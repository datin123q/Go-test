// cmd/codegen.go
package main

import (
	"fmt"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/genhelper"
	"github.com/spf13/cobra"
)

var codegenCmd = &cobra.Command{
	Use:   "codegen",
	Short: "도메인 코드를 생성합니다",
	Long:  `config/code 디렉토리의 yaml 파일들을 기반으로 도메인 코드를 생성합니다.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputDir := "config/code"
		outputDir := "domain/codegen"

		fmt.Printf("Code generation started...\n")
		fmt.Printf("Input directory: %s\n", inputDir)
		fmt.Printf("Output directory: %s\n", outputDir)

		generator, err := genhelper.NewGenerator(inputDir, outputDir)
		if err != nil {
			fmt.Printf("Failed to create generator: %v\n", err)
			return
		}

		if err := generator.Generate(); err != nil {
			fmt.Printf("Failed to generate code: %v\n", err)
			return
		}

		fmt.Println("Code generation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(codegenCmd)
}
