package csvutils

import (
	"encoding/csv"
	"mime/multipart"
	"net/http"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func ReadMultpartCsv(file *multipart.File) ([][]string, error) {

	// Open CSV file
	// Read File into a Variable

	lines, err := csv.NewReader(transform.NewReader(*file, japanese.ShiftJIS.NewDecoder())).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}

func ReadFileCsv(file *os.File) ([][]string, error) {

	// Open CSV file
	// Read File into a Variable

	lines, err := csv.NewReader(transform.NewReader(file, japanese.ShiftJIS.NewDecoder())).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}

func WriteHeader(w http.ResponseWriter, filename string) {
	w.Write([]byte{0xEF, 0xBB, 0xBF})
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
}

func WriteHeaderWithShiftJIS(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "text/csv; charset=Shift_JIS")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
}
