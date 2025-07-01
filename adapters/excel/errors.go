package excel

import "errors"

var (
	// ErrMissingFilePath is returned when file path is not specified
	ErrMissingFilePath = errors.New("file path is required")

	// ErrMissingSheetName is returned when sheet name is not specified
	ErrMissingSheetName = errors.New("sheet name is required")

	// ErrSheetNotFound is returned when the specified sheet doesn't exist
	ErrSheetNotFound = errors.New("sheet not found")

	// ErrInvalidFileFormat is returned when the file is not a valid Excel file
	ErrInvalidFileFormat = errors.New("invalid Excel file format")
)
