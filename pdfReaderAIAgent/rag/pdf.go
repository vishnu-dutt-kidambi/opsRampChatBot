package rag

import (
	"bytes"
	"fmt"

	"github.com/ledongthuc/pdf"
)

// =============================================================================
// PDF Text Extraction
// =============================================================================
//
// WHAT HAPPENS HERE:
// PDFs are complex documents that can contain text, images, fonts, and more.
// This function extracts the plain text content from a PDF file.
//
// LIMITATIONS:
// - Only works with text-based PDFs (not scanned images)
// - Scanned PDFs would need OCR (Optical Character Recognition)
// - Complex layouts (tables, columns) may not extract perfectly
//
// For production use, consider libraries like:
//   - unidoc/unipdf (Go, commercial)
//   - Apache Tika (Java, open source)
//   - pdfplumber (Python, great for tables)
// =============================================================================

// ExtractTextFromPDF reads a PDF file and returns all text content.
// Returns the extracted text, page count, and any error.
func ExtractTextFromPDF(filePath string) (string, int, error) {
	// Open the PDF file for reading
	f, reader, err := pdf.Open(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("cannot open PDF '%s': %w", filePath, err)
	}
	defer f.Close()

	// Get the total number of pages
	pageCount := reader.NumPage()

	// Extract plain text from ALL pages at once
	// The library handles page ordering and text flow
	var buf bytes.Buffer
	plainText, err := reader.GetPlainText()
	if err != nil {
		return "", 0, fmt.Errorf("cannot extract text from PDF: %w", err)
	}

	// Read all the extracted text into our buffer
	_, err = buf.ReadFrom(plainText)
	if err != nil {
		return "", 0, fmt.Errorf("cannot read extracted text: %w", err)
	}

	return buf.String(), pageCount, nil
}
