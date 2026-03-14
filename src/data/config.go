package data

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	GoogleSheetDataSource []GoogleSheetDataSource `mapstructure:"google_sheet_data_sources"`
}

type GoogleSheetDataSource struct {
	SpreadSheetID       string `mapstructure:"spreadsheet_id"`
	CredentialsFilePath string `mapstructure:"credentials_file_path"`
}

func (gsds *GoogleSheetDataSource) Validate() error {
	if gsds.SpreadSheetID == "" {
		return fmt.Errorf("spreadsheet_id is required")
	}
	if gsds.CredentialsFilePath == "" {
		return fmt.Errorf("credentials_file_path is required")
	}

	if _, err := filepath.Abs(gsds.CredentialsFilePath); err != nil {
		return fmt.Errorf("invalid credentials_file_path: %w", err)
	}

	if _, err := os.Stat(gsds.CredentialsFilePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("credentials_file_path does not exist: %s", gsds.CredentialsFilePath)
		}
		return fmt.Errorf("error checking credentials_file_path: %w", err)
	}

	return nil
}
