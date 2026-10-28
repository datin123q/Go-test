package genhelper

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/genhelper/code_template"
)

type Generator struct {
	InputDir  string
	OutputDir string
	tmpl      *template.Template
}

func NewGenerator(inputDir, outputDir string) (*Generator, error) {
	funcMap := template.FuncMap{
		"toCamel": toCamelCase,
		"toLower": toLower,
	}

	tmpl, err := template.New("code").Funcs(funcMap).Parse(code_template.Domain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse code_template")
	}

	return &Generator{
		InputDir:  inputDir,
		OutputDir: outputDir,
		tmpl:      tmpl,
	}, nil
}

func (g *Generator) Generate() error {
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create output directory")
	}

	// Create common code file
	commonFile := filepath.Join(g.OutputDir, "common.go")
	if err := os.WriteFile(commonFile, []byte(code_template.Common), 0644); err != nil {
		return errors.Wrap(err, "failed to create common file")
	}

	// 各ドメインディレクトリを走査する
	domains, err := os.ReadDir(g.InputDir)
	if err != nil {
		return errors.Wrap(err, "failed to read input directory")
	}

	for _, domain := range domains {
		if !domain.IsDir() {
			continue
		}

		// 各ドメインのcode.yamlファイルを処理する
		yamlPath := filepath.Join(g.InputDir, domain.Name(), "code.yaml")
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			continue // code.yamlがないディレクトリはスキップします
		}

		config, err := g.loadYAML(yamlPath)
		if err != nil {
			return errors.Wrapf(err, "failed to load %s", yamlPath)
		}

		// コード生成
		if err := g.generateDomainFile(domain.Name(), config); err != nil {
			return errors.Wrapf(err, "failed to generate code for %s", domain.Name())
		}
	}

	return nil
}

func (g *Generator) generateDomainFile(domain string, config DomainConfig) error {
	outFile := filepath.Join(g.OutputDir, strings.ToLower(domain)+".go")
	f, err := os.Create(outFile)
	if err != nil {
		return errors.Wrap(err, "failed to create output file")
	}
	defer f.Close()

	err = g.tmpl.Execute(f, struct {
		Domain string
		Config DomainConfig
	}{
		Domain: strings.ToUpper(domain),
		Config: config,
	})

	if err != nil {
		return errors.Wrap(err, "failed to generate code")
	}

	return nil
}

// loadYAMLはYAMLファイルを読み込んでDomainConfigに変換します
func (g *Generator) loadYAML(path string) (DomainConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read YAML file")
	}

	var config DomainConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "failed to parse YAML")
	}

	return config, nil
}

// toCamelCaseはSNAKE_CASEをCamelCaseに変換します
func toCamelCase(s string) string {
	// 空の文字列処理
	if s == "" {
		return s
	}

	// 既にキャメルケースの場合はそのまま返す
	// (アンダースコアがなく、最初の文字が小文字で、大文字が1つ以上含まれている場合)
	if !strings.Contains(s, "_") &&
		unicode.IsLower([]rune(s)[0]) &&
		s != strings.ToLower(s) {
		return s
	}

	// すべての入力を小文字に変換
	s = strings.ToLower(s)

	words := strings.Split(s, "_")
	var result strings.Builder

	for i, word := range words {
		if word == "" {
			continue
		}

		if i == 0 && strings.HasPrefix(s, "_") {
			// アンダースコアで始まる場合、最初の単語も大文字にする
			result.WriteString(cases.Title(language.English).String(word))
		} else if i == 0 {
			// 最初の単語は小文字で
			result.WriteString(word)
		} else {
			// その他の単語は大文字で始まる
			result.WriteString(cases.Title(language.English).String(word))
		}
	}

	return result.String()
}

// toLowerは文字列を小文字に変換します
func toLower(s string) string {
	return strings.ToLower(s)
}
