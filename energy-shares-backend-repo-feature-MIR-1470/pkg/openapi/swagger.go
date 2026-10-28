package openapi

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"

	"github.com/swaggest/openapi-go/openapi31"
)

type SwaggerUISetup struct {
	cfg     *config.Config
	openapi *openapi31.Reflector
}

func NewSwaggerUISetup(cfg *config.Config, openapi *openapi31.Reflector) *SwaggerUISetup {
	return &SwaggerUISetup{cfg: cfg, openapi: openapi}
}

// Setup は Swagger UI のアセットを書き込み可能な一時ディレクトリにダウンロード・配置します。
// 呼び出し元が静的ファイルを配信できるよう、配置先のディレクトリパスを返します。
func (s *SwaggerUISetup) Setup() (string, error) {
	// os.TempDir()（通常 /tmp）を使用することで、読み取り専用のコンテナファイルシステム
	//（例: AWS ECS の read-only ルートFS）でのパーミッションエラーを回避します。
	tmpDir := os.TempDir()
	zipPath := filepath.Join(tmpDir, "swagger-ui.zip")
	swaggerUIDir := filepath.Join(tmpDir, "swagger-ui")
	extractDir := filepath.Join(tmpDir, fmt.Sprintf("swagger-ui-%s", s.cfg.Swagger.Version))

	downloadUrl := fmt.Sprintf("https://github.com/swagger-api/swagger-ui/archive/refs/tags/v%s.zip", s.cfg.Swagger.Version)
	err := downloadFile(downloadUrl, zipPath)
	if err != nil {
		return "", errors.Wrap(err, "error downloading Swagger UI")
	}
	defer os.Remove(zipPath)

	err = unzip(zipPath, tmpDir)
	if err != nil {
		return "", errors.Wrap(err, "error unzipping Swagger UI")
	}

	sourceDir := filepath.Join(extractDir, "dist")
	filesToCopy := []string{
		"swagger-ui.css",
		"swagger-ui-bundle.js",
		"swagger-ui-standalone-preset.js",
		"swagger-initializer.js",
		"index.html",
		"index.css",
	}

	err = os.MkdirAll(swaggerUIDir, os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "error creating swagger-ui directory")
	}

	for _, file := range filesToCopy {
		if err = copyFile(filepath.Join(sourceDir, file), filepath.Join(swaggerUIDir, file)); err != nil {
			return "", errors.Wrap(err, "error copying file")
		}
	}

	// Generate and save OpenAPI schema
	schema, err := s.openapi.Spec.MarshalYAML()
	if err != nil {
		return "", errors.Wrap(err, "error marshaling OpenAPI schema")
	}

	err = os.WriteFile(filepath.Join(swaggerUIDir, "openapi.yaml"), schema, 0644)
	if err != nil {
		return "", errors.Wrap(err, "error writing OpenAPI schema")
	}

	err = modifySwaggerInitializer(filepath.Join(swaggerUIDir, "swagger-initializer.js"), s.cfg.Swagger.OpenApiSpecPath)
	if err != nil {
		return "", errors.Wrap(err, "error modifying swagger-initializer.js")
	}

	err = os.RemoveAll(extractDir)
	if err != nil {
		return "", errors.Wrap(err, "error removing temporary directory")
	}

	return swaggerUIDir, nil
}

func downloadFile(url, filepath string) error {
	file, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "error downloading file")
	}
	defer file.Body.Close()

	if file.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected HTTP status %d when downloading %s", file.StatusCode, url)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return errors.Wrap(err, "error creating file")
	}
	defer out.Close()

	if _, err = io.Copy(out, file.Body); err != nil {
		return errors.Wrap(err, "error copying file")
	}

	return nil
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return errors.Wrap(err, "error opening zip file")
	}

	defer func(r *zip.ReadCloser) {
		_ = r.Close()
	}(r)

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return errors.Wrap(err, "error creating directory")
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return errors.Wrap(err, "error opening file")
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return errors.Wrap(err, "error opening file")
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return errors.Wrap(err, "error copying file")
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "error opening source file")
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return errors.Wrap(err, "error creating destination file")
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return errors.Wrap(err, "error copying file")
	}

	if err = out.Close(); err != nil {
		return errors.Wrap(err, "error closing destination file")
	}

	return nil
}

func modifySwaggerInitializer(filepath string, openAPISpecPath string) error {
	content := fmt.Sprintf(
		`
window.onload = function() {
    function configureSwaggerUI(options) {
        const defaultOptions = {
            url: "%s",
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl,
                SwaggerUIBundle.plugins.TopBar,
                SwaggerUIBundle.plugins.TryItOutPlugin
            ],
            layout: "StandaloneLayout",
            filter: true,
            syntaxHighlight: {
                activated: true,
                theme: "agate"
            },
            displayRequestDuration: true,
            defaultModelsExpandDepth: -1,
            defaultModelExpandDepth: 1,
            docExpansion: 'list',
            showExtensions: true,
            showCommonExtensions: true,
            tryItOutEnabled: false,
            persistAuthorization: true,
            tagsSorter: 'alpha',
            operationsSorter: 'alpha',
            onComplete: function() {
                // Add copy button for cURL commands
                function addCopyButtons() {
                    document.querySelectorAll('.opblock').forEach(opblock => {
                        const curlSpan = opblock.querySelector('.curl');
                        if (curlSpan && !curlSpan.nextElementSibling?.classList.contains('copy-curl-button')) {
                            const copyButton = document.createElement('button');
                            copyButton.textContent = 'Copy cURL';
                            copyButton.className = 'copy-curl-button';
                            copyButton.style.cssText = 'margin-left: 10px; padding: 5px 10px; border-radius: 4px; background-color: #4990e2; color: white; border: none;';
                            copyButton.onclick = function() {
                                navigator.clipboard.writeText(curlSpan.textContent).then(() => {
                                    copyButton.textContent = 'Copied!';
                                    setTimeout(() => copyButton.textContent = 'Copy cURL', 2000);
                                });
                            };
                            curlSpan.parentNode.insertBefore(copyButton, curlSpan.nextSibling);
                        }
                    });
                }

                // Initial addition of copy buttons
                addCopyButtons();

                // Set up a MutationObserver to watch for changes in the DOM
                const observer = new MutationObserver(mutations => {
                    mutations.forEach(mutation => {
                        if (mutation.type === 'childList') {
                            addCopyButtons();
                        }
                    });
                });

                // Start observing the document with the configured parameters
                observer.observe(document.body, { childList: true, subtree: true });
            }
        };

        const mergedOptions = Object.assign({}, defaultOptions, options);

        // Initialize Swagger UI
        const ui = SwaggerUIBundle(mergedOptions);

        return ui;
    }

    const ui = configureSwaggerUI({
        validatorUrl: "https://validator.swagger.io/validator",
        oauth2RedirectUrl: `+"`${window.location.protocol}//${window.location.host}/oauth2-redirect.html`"+`
    });
};
`,
		openAPISpecPath)

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return errors.Wrap(err, "error writing to file")
	}

	return nil
}
