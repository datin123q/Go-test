package swagger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/sirupsen/logrus"
)

// NullableProperty identifies a Swagger 2.0 schema property that accepts null.
type NullableProperty struct {
	Definition string
	Property   string
}

// GenerateSwaggerIfNeeded checks the running mode and initiates
// the Swagger JSON generation process if required.
func GenerateSwaggerIfNeeded(
	apiRoot echoswagger.ApiRoot,
	echoInstance *echo.Echo,
	baseURL, saveFile string,
	nullableProperties ...NullableProperty,
) {
	swaggerMode := os.Getenv("SWAGGER_MODE")
	if swaggerMode != "GEN_SWAGGER_MODE" {
		return
	}

	swaggerJSON, err := generateSwaggerJSON(apiRoot, echoInstance, baseURL)
	if err != nil {
		logrus.Fatalf("Failed to generate Swagger JSON: %s", err)
	}
	swaggerJSON, err = addNullableProperties(swaggerJSON, nullableProperties)
	if err != nil {
		logrus.Fatalf("Failed to add nullable properties to Swagger JSON: %s", err)
	}

	gitRoot := os.Getenv("GIT_ROOT")
	if gitRoot == "" {
		gitRoot = "./"
	}
	savePath := filepath.Join(gitRoot, "swagger", saveFile)
	if err := saveSwaggerJSON(swaggerJSON, savePath); err != nil {
		logrus.Fatalf("Failed to save Swagger JSON: %s", err)
	}

	fmt.Printf("%v: created and saved successfully.\n", savePath)
	os.Exit(0)
}

func addNullableProperties(swaggerJSON []byte, nullableProperties []NullableProperty) ([]byte, error) {
	if len(nullableProperties) == 0 {
		return swaggerJSON, nil
	}

	var document map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(swaggerJSON))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}

	definitions, ok := document["definitions"].(map[string]interface{})
	if !ok {
		return nil, errors.New("swagger definitions not found")
	}

	for _, nullableProperty := range nullableProperties {
		definition, ok := definitions[nullableProperty.Definition].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("swagger definition not found: %s", nullableProperty.Definition)
		}

		properties, ok := definition["properties"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("swagger properties not found: %s", nullableProperty.Definition)
		}

		property, ok := properties[nullableProperty.Property].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf(
				"swagger property not found: %s.%s",
				nullableProperty.Definition,
				nullableProperty.Property,
			)
		}
		property["x-nullable"] = true
	}

	return json.MarshalIndent(document, "", "  ")
}

// generateSwaggerJSON creates the Swagger JSON and returns it.
func generateSwaggerJSON(apiRoot echoswagger.ApiRoot, echoInstance *echo.Echo, baseURL string) ([]byte, error) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	context := echoInstance.NewContext(request, recorder)

	root := apiRoot.(*echoswagger.Root)
	swaggerSpec, err := root.GetSpec(context, "/doc")
	if err != nil {
		return nil, err
	}
	swaggerSpec.Host = baseURL
	swaggerSpec.Schemes = []string{"https"}

	swaggerJSON, err := json.MarshalIndent(swaggerSpec, "", "  ")
	if err != nil {
		return nil, err
	}

	return swaggerJSON, nil
}

// saveSwaggerJSON saves the given Swagger JSON to the specified path.
func saveSwaggerJSON(swaggerJSON []byte, savePath string) error {
	// generate dir if not exists
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(savePath, swaggerJSON, 0644)
}
