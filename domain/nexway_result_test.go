package domain

import (
	"encoding/json"
	"testing"
)

func TestNexwayImageType_Valid(t *testing.T) {
	tests := []struct {
		name  string
		it    NexwayImageType
		valid bool
	}{
		{"front is valid", ImageTypeFront, true},
		{"back is valid", ImageTypeBack, true},
		{"diagonal is valid", ImageTypeDiagonal, true},
		{"front_trim is valid", ImageTypeFrontTrim, true},
		{"back_trim is valid", ImageTypeBackTrim, true},
		{"diagonal_trim is valid", ImageTypeDiagonalTrim, true},
		{"selfie is valid", ImageTypeSelfie, true},
		{"invalid is invalid", NexwayImageType("invalid"), false},
		{"empty is invalid", NexwayImageType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.it.Valid(); got != tt.valid {
				t.Errorf("NexwayImageType.Valid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestNexwayImageType_IsTrimmed(t *testing.T) {
	tests := []struct {
		name    string
		it      NexwayImageType
		trimmed bool
	}{
		{"front is not trimmed", ImageTypeFront, false},
		{"front_trim is trimmed", ImageTypeFrontTrim, true},
		{"back_trim is trimmed", ImageTypeBackTrim, true},
		{"diagonal_trim is trimmed", ImageTypeDiagonalTrim, true},
		{"selfie is not trimmed", ImageTypeSelfie, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.it.IsTrimmed(); got != tt.trimmed {
				t.Errorf("NexwayImageType.IsTrimmed() = %v, want %v", got, tt.trimmed)
			}
		})
	}
}

func TestNexwayImageType_IsDocument(t *testing.T) {
	tests := []struct {
		name     string
		it       NexwayImageType
		document bool
	}{
		{"front is document", ImageTypeFront, true},
		{"back is document", ImageTypeBack, true},
		{"diagonal is document", ImageTypeDiagonal, true},
		{"selfie is not document", ImageTypeSelfie, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.it.IsDocument(); got != tt.document {
				t.Errorf("NexwayImageType.IsDocument() = %v, want %v", got, tt.document)
			}
		})
	}
}

func TestNexwayImageType_RequiresFaceMatch(t *testing.T) {
	tests := []struct {
		name      string
		it        NexwayImageType
		faceMatch bool
	}{
		{"front does not require face match", ImageTypeFront, false},
		{"selfie requires face match", ImageTypeSelfie, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.it.RequiresFaceMatch(); got != tt.faceMatch {
				t.Errorf("NexwayImageType.RequiresFaceMatch() = %v, want %v", got, tt.faceMatch)
			}
		})
	}
}

func TestNexwayImageType_JSON_Unmarshal(t *testing.T) {
	// JSONからGo構造体への変換テスト
	jsonData := `{"image_type":"front","image_data":"base64data"}`

	type TestImage struct {
		ImageType NexwayImageType `json:"image_type"`
		ImageData string          `json:"image_data"`
	}

	var img TestImage
	if err := json.Unmarshal([]byte(jsonData), &img); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if img.ImageType != ImageTypeFront {
		t.Errorf("NexwayImageType = %v, want %v", img.ImageType, ImageTypeFront)
	}

	if !img.ImageType.Valid() {
		t.Error("NexwayImageType should be valid")
	}
}

func TestNexwayImageType_JSON_Marshal(t *testing.T) {
	// Go構造体からJSONへの変換テスト
	type TestImage struct {
		ImageType NexwayImageType `json:"image_type"`
		ImageData string          `json:"image_data"`
	}

	img := TestImage{
		ImageType: ImageTypeSelfie,
		ImageData: "base64data",
	}

	jsonData, err := json.Marshal(img)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	expected := `{"image_type":"selfie","image_data":"base64data"}`
	if string(jsonData) != expected {
		t.Errorf("json.Marshal() = %s, want %s", string(jsonData), expected)
	}
}
