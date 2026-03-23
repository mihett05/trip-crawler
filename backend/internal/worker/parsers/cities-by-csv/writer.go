package city

import (
	"encoding/json"
	"os"
)

// SaveJSON сохраняет данные в файл в формате JSON (с форматированием)
func SaveJSON(data any, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")

	return encoder.Encode(data)
}
