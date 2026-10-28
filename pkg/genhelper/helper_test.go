package genhelper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name      string
		inputDir  string
		outputDir string
		wantErr   bool
	}{
		{
			name:      "유효한 디렉토리 경로",
			inputDir:  "testdata/input",
			outputDir: "testdata/output",
			wantErr:   false,
		},
		{
			name:      "빈 입력 디렉토리",
			inputDir:  "",
			outputDir: "testdata/output",
			wantErr:   false,
		},
		{
			name:      "빈 출력 디렉토리",
			inputDir:  "testdata/input",
			outputDir: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewGenerator(tt.inputDir, tt.outputDir)
			assert.Equal(t, tt.wantErr, err != nil)

			if tt.wantErr {
				assert.Nil(t, generator)
				return
			}

			assert.NotNil(t, generator)
			assert.Equal(t, tt.inputDir, generator.InputDir)
			assert.Equal(t, tt.outputDir, generator.OutputDir)
			assert.NotNil(t, generator.tmpl)
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputDir := filepath.Join(tmpDir, "output")

	validYAML := `USER:
  status:
    codes:
      - code: ACTIVE
        descriptions:
          ko: "활성 상태"
          en: "Active status"
      - code: INACTIVE
        descriptions:
          ko: "비활성 상태"
          en: "Inactive status"`

	invalidYAML := `USER:
  status:
    codes: [invalid`

	tests := []struct {
		name    string
		yaml    string
		domain  string
		wantErr bool
	}{
		{
			name:    "정상적인 코드 생성",
			yaml:    validYAML,
			domain:  "user",
			wantErr: false,
		},
		{
			name:    "잘못된 YAML 파일",
			yaml:    invalidYAML,
			domain:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDomainDir := filepath.Join(inputDir, tt.domain)
			err := os.MkdirAll(testDomainDir, 0755)
			assert.NoError(t, err)

			err = os.MkdirAll(outputDir, 0755)
			assert.NoError(t, err)

			yamlPath := filepath.Join(testDomainDir, "code.yaml")
			err = os.WriteFile(yamlPath, []byte(tt.yaml), 0644)
			assert.NoError(t, err)

			generator, err := NewGenerator(inputDir, outputDir)
			assert.NoError(t, err)

			err = generator.Generate()
			assert.Equal(t, tt.wantErr, err != nil)

			if tt.wantErr {
				return
			}

			content, err := os.ReadFile(filepath.Join(outputDir, "user.go"))
			assert.NoError(t, err)

			generatedCode := string(content)
			assert.Contains(t, generatedCode, "type user_statusCode string")
			assert.Contains(t, generatedCode, "USER.status")
			assert.Contains(t, generatedCode, "ACTIVE")
			assert.Contains(t, generatedCode, "INACTIVE")
			assert.Contains(t, generatedCode, "활성 상태")
			assert.Contains(t, generatedCode, "비활성 상태")
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "스네이크 케이스를 카멜 케이스로 변환",
			input: "hello_world",
			want:  "helloWorld",
		},
		{
			name:  "여러 개의 언더스코어가 있는 경우",
			input: "hello_wonderful_world",
			want:  "helloWonderfulWorld",
		},
		{
			name:  "이미 카멜케이스인 경우 유지",
			input: "helloWorld",
			want:  "helloWorld",
		},
		{
			name:  "단일 단어는 소문자 유지",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "빈 문자열 처리",
			input: "",
			want:  "",
		},
		{
			name:  "숫자가 포함된 경우",
			input: "hello_world_123",
			want:  "helloWorld123",
		},
		{
			name:  "여러 개의 연속된 언더스코어 처리",
			input: "hello___world",
			want:  "helloWorld",
		},
		{
			name:  "모두 대문자인 경우",
			input: "HELLO_WORLD",
			want:  "helloWorld",
		},
		{
			name:  "대소문자가 섞여있는 경우",
			input: "Hello_World",
			want:  "helloWorld",
		},
		{
			name:  "언더스코어로 시작하는 경우",
			input: "_hello_world",
			want:  "HelloWorld",
		},
		{
			name:  "언더스코어로 끝나는 경우",
			input: "hello_world_",
			want:  "helloWorld",
		},
		{
			name:  "대문자의 경우",
			input: "HELLO",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toCamelCase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadYAML(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		validate    func(*testing.T, DomainConfig)
	}{
		{
			name: "정상적인 YAML 파일",
			yamlContent: `
USER:
 status:
   codes:
     - code: ACTIVE
       descriptions:
         ko: "활성 상태"
         en: "Active status"
`,
			wantErr: false,
			validate: func(t *testing.T, config DomainConfig) {
				assert.Contains(t, config, "USER")
				assert.Contains(t, config["USER"], "status")
				assert.Len(t, config["USER"]["status"].Codes, 1)
				assert.Equal(t, "ACTIVE", config["USER"]["status"].Codes[0].Code)
			},
		},
		{
			name: "잘못된 YAML 문법",
			yamlContent: `
USER:
 status:
   codes:
     - code: ACTIVE
       descriptions:
         ko: "활성 상태
         en: "Active status"
`,
			wantErr: true,
		},
		{
			name:        "빈 YAML 파일",
			yamlContent: "",
			wantErr:     false,
			validate: func(t *testing.T, config DomainConfig) {
				assert.Empty(t, config)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.yaml")
			err := os.WriteFile(testFile, []byte(tt.yamlContent), 0644)
			assert.NoError(t, err)

			generator := &Generator{
				InputDir: tmpDir,
			}

			config, err := generator.loadYAML(testFile)
			assert.Equal(t, tt.wantErr, err != nil)

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}
