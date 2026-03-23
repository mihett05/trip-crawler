package utils

import (
	"encoding/json"
	"os"
)

// SaveJSON сохраняет любые данные в файл.
func SaveJSON(data any, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	encoder.SetEscapeHTML(false)

	return encoder.Encode(data)
}
