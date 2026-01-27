// pkg/aip/processor_test.go
package aip

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadAndSaveAIP(t *testing.T) {
	// 1. Setup paths for the test data and a temporary output file.
	inputPath := filepath.Join("testdata", "sample.aip")
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "output.aip")

	// 2. Load the original AIP file.
	originalDoc, err := LoadAIP(inputPath)
	if err != nil {
		t.Fatalf("Failed to load AIP file: %v", err)
	}

	// 3. Save the loaded structure to the temporary output file.
	err = SaveAIP(originalDoc, outputPath)
	if err != nil {
		t.Fatalf("Failed to save AIP file: %v", err)
	}

	// 4. Load the newly saved file.
	savedDoc, err := LoadAIP(outputPath)
	if err != nil {
		t.Fatalf("Failed to load the newly saved AIP file: %v", err)
	}

	// 5. Compare the original structure with the re-loaded one to ensure no data was lost.
	if !reflect.DeepEqual(originalDoc, savedDoc) {
		t.Errorf("The saved and re-loaded AIP document does not match the original.")
	}
}

// createTestDocument is a generic helper to create a Document struct for testing.
func createTestDocument(sourcePaths ...string) *Document {
	doc := &Document{
		Components: []Component{
			{
				CID:  "caphyon.advinst.msicomp.MsiFilesComponent",
				Rows: []Row{},
			},
		},
	}
	for _, p := range sourcePaths {
		doc.Components[0].Rows = append(doc.Components[0].Rows, Row{SourcePath: p})
	}
	return doc
}

// --- NEW, MORE COMPREHENSIVE TEST ---
func TestResolveSourcePathsWithBuildType(t *testing.T) {
	// 1. Setup: Create a temporary directory structure.
	searchDir := t.TempDir()
	commonDir := filepath.Join(searchDir, "Common")
	packedDir := filepath.Join(searchDir, "Packed")
	unpackedDir := filepath.Join(searchDir, "Unpacked")
	os.MkdirAll(commonDir, 0755)
	os.MkdirAll(packedDir, 0755)
	os.MkdirAll(unpackedDir, 0755)

	// Create dummy files. Note that 'app.exe' exists in two places.
	commonFilePath := filepath.Join(commonDir, "common.exe")
	packedFilePath := filepath.Join(packedDir, "app.exe")
	unpackedFilePath := filepath.Join(unpackedDir, "app.exe")
	os.WriteFile(commonFilePath, []byte("c"), 0644)
	os.WriteFile(packedFilePath, []byte("p"), 0644)
	os.WriteFile(unpackedFilePath, []byte("u"), 0644)

	// 2. Arrange: Create a document that references both files.
	doc := createTestDocument(`C:\old\path\to\app.exe`, `C:\another\path\common.exe`)

	t.Run("Resolves using Packed path for packed buildType", func(t *testing.T) {
		// Act: Run the resolver for a "packed" build.
		_, err := doc.ResolveSourcePaths(searchDir, "packed")
		if err != nil {
			t.Fatalf("ResolveSourcePaths() returned an unexpected error: %v", err)
		}

		// Assert: Check that app.exe resolved to the Packed directory.
		resolvedAppPath := doc.Components[0].Rows[0].SourcePath
		if resolvedAppPath != packedFilePath {
			t.Errorf("Expected app.exe path to be '%s', but got '%s'", packedFilePath, resolvedAppPath)
		}
		resolvedCommonPath := doc.Components[0].Rows[1].SourcePath
		if resolvedCommonPath != commonFilePath {
			t.Errorf("Expected common.exe path to be '%s', but got '%s'", commonFilePath, resolvedCommonPath)
		}
	})

	t.Run("Resolves using Normal path for normal buildType", func(t *testing.T) {
		// Act: Run the resolver for an "normal" build.
		_, err := doc.ResolveSourcePaths(searchDir, "normal")
		if err != nil {
			t.Fatalf("ResolveSourcePaths() returned an unexpected error: %v", err)
		}

		// Assert: Check that app.exe resolved to the Unpacked directory.
		resolvedAppPath := doc.Components[0].Rows[0].SourcePath
		if resolvedAppPath != unpackedFilePath {
			t.Errorf("Expected app.exe path to be '%s', but got '%s'", unpackedFilePath, resolvedAppPath)
		}
		resolvedCommonPath := doc.Components[0].Rows[1].SourcePath
		if resolvedCommonPath != commonFilePath {
			t.Errorf("Expected common.exe path to be '%s', but got '%s'", commonFilePath, resolvedCommonPath)
		}
	})

	t.Run("Returns error for missing files", func(t *testing.T) {
		// Arrange: Create a document that references a file that does not exist.
		missingFileName := "missing_file.dll"
		docWithMissing := createTestDocument(missingFileName)

		// Act & Assert
		_, err := docWithMissing.ResolveSourcePaths(searchDir, "packed")
		if err == nil {
			t.Fatal("Expected an error for a missing file, but got nil")
		}
		if !strings.Contains(err.Error(), missingFileName) {
			t.Errorf("Expected error message to contain '%s', but got: %v", missingFileName, err)
		}
	})
}
