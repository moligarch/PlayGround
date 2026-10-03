// pkg/aip/processor.go
package aip

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LoadAIP reads and parses an .aip XML file from the given path.
func LoadAIP(path string) (*Document, error) {
	xmlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc Document
	if err := xml.Unmarshal(xmlFile, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// SetRegistryValue finds a registry row by its 'Name' and updates its 'Value'.
func (d *Document) SetRegistryValue(name, value string) error {
	for i, component := range d.Components {
		if component.CID == "caphyon.advinst.msicomp.MsiRegsComponent" {
			for j, row := range component.Rows {
				if row.Name == name {
					d.Components[i].Rows[j].Value = value
					return nil
				}
			}
		}
	}
	return fmt.Errorf("registry row with Name='%s' not found", name)
}

// SetProperty finds a property row by its 'Property' name and updates its 'Value'.
func (d *Document) SetProperty(name, value string) error {
	for i, component := range d.Components {
		if component.CID == "caphyon.advinst.msicomp.MsiPropsComponent" {
			for j, row := range component.Rows {
				if row.Property == name {
					d.Components[i].Rows[j].Value = value
					return nil
				}
			}
		}
	}
	return fmt.Errorf("property row with Property='%s' not found", name)
}

// SetProperty finds a property row by its 'Property' name and updates its 'Value'.
func (d *Document) GetProperty(name string) (string, error) {
	for i, component := range d.Components {
		if component.CID == "caphyon.advinst.msicomp.MsiPropsComponent" {
			for j, row := range component.Rows {
				if row.Property == name {
					return d.Components[i].Rows[j].Value, nil
				}
			}
		}
	}
	return "", fmt.Errorf("property row with Property='%s' not found", name)
}

type FileIndex struct {
	// Map filename to a slice of ALL absolute paths that share that filename
	CommonFiles   map[string][]string
	PackedFiles   map[string][]string
	UnpackedFiles map[string][]string
}

// buildFileIndex walks the search directories and creates a detailed index of all files.
func buildFileIndex(baseDir string) (*FileIndex, error) {
	index := &FileIndex{
		CommonFiles:   make(map[string][]string),
		PackedFiles:   make(map[string][]string),
		UnpackedFiles: make(map[string][]string),
	}

	indexDir := func(dirName string, targetMap map[string][]string) error {
		targetDir := filepath.Join(baseDir, dirName)
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			return nil // Skip if directory doesn't exist
		}

		return filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				fileName := d.Name()
				// Append to slice instead of overwriting!
				targetMap[fileName] = append(targetMap[fileName], filepath.ToSlash(path))
			}
			return nil
		})
	}

	if err := indexDir("Common", index.CommonFiles); err != nil {
		return nil, err
	}
	if err := indexDir("Packed", index.PackedFiles); err != nil {
		return nil, err
	}
	if err := indexDir("Unpacked", index.UnpackedFiles); err != nil {
		return nil, err
	}

	return index, nil
}

// findBestMatch finds the physical file path that best matches the directory structure of the AIP path.
func findBestMatch(aipSourcePath string, availablePaths []string) (string, bool) {
	if len(availablePaths) == 0 {
		return "", false
	}
	if len(availablePaths) == 1 {
		return availablePaths[0], true // Only one option, return it immediately
	}

	aipParts := strings.Split(filepath.ToSlash(aipSourcePath), "/")
	
	bestMatch := ""
	maxScore := -1

	for _, physPath := range availablePaths {
		physParts := strings.Split(physPath, "/")
		score := 0
		
		// Work backwards checking how many parent directories match
		aipIdx := len(aipParts) - 1
		physIdx := len(physParts) - 1
		
		for aipIdx >= 0 && physIdx >= 0 {
			if strings.EqualFold(aipParts[aipIdx], physParts[physIdx]) {
				score++
			} else {
				break
			}
			aipIdx--
			physIdx--
		}

		if score > maxScore {
			maxScore = score
			bestMatch = physPath
		}
	}
	
	return bestMatch, true
}

func (d *Document) ResolveSourcePaths(searchDir, buildType string) (int, error) {
	index, err := buildFileIndex(searchDir)
	if err != nil {
		return 0, err
	}

	changedCount := 0
	var missingFiles []string

	for i, component := range d.Components {
		if component.CID == "caphyon.advinst.msicomp.MsiFilesComponent" ||
			component.CID == "caphyon.advinst.msicomp.MsiIconsComponent" {
			for j, row := range component.Rows {
				if row.SourcePath == "" {
					continue
				}

				aipPath := filepath.ToSlash(row.SourcePath)
				parts := strings.Split(aipPath, "/")
				fileName := parts[len(parts)-1]

				var absolutePath string
				var found bool

				// 1. Check Packed/Unpacked first based on build type
				targetMap := index.UnpackedFiles
				if buildType == "packed" {
					targetMap = index.PackedFiles
				}

				if paths, ok := targetMap[fileName]; ok {
					absolutePath, found = findBestMatch(aipPath, paths)
				}

				// 2. Fall back to Common files if not found
				if !found {
					if paths, ok := index.CommonFiles[fileName]; ok {
						absolutePath, found = findBestMatch(aipPath, paths)
					}
				}

				if found {
					d.Components[i].Rows[j].SourcePath = absolutePath
					changedCount++
				} else {
					missingFiles = append(missingFiles, fileName)
				}
			}
		}
	}

	if len(missingFiles) > 0 {
		return changedCount, fmt.Errorf("the following resources were not found for a '%s' build:\n- %s", buildType, strings.Join(missingFiles, "\n- "))
	}
	return changedCount, nil
}

// SetProductVersion finds the ProductVersion property and updates it.
func (d *Document) SetProductVersion(version string) error {
	return d.SetProperty("ProductVersion", version)
}

// SetBuildOutput finds the build row and updates the output MSI folder and filename.
func (d *Document) SetBuildOutput(folder, fileName string) error {
	for i, component := range d.Components {
		if component.CID == "caphyon.advinst.msicomp.BuildComponent" {
			for j := range component.Rows {
				if component.Rows[j].BuildKey == "DefaultBuild" {
					d.Components[i].Rows[j].PackageFolder = folder
					d.Components[i].Rows[j].PackageFileName = fileName
					return nil
				}
			}
		}
	}
	return fmt.Errorf("build component row with BuildKey='DefaultBuild' not found")
}

func GenerateAIPBytes(doc *Document) ([]byte, error) {
	bytes, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	header := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	finalBytes := append(header, bytes...)
	return finalBytes, nil
}

// SaveAIP is now a simple helper that generates bytes and writes them to a file.
func SaveAIP(doc *Document, path string) error {
	finalBytes, err := GenerateAIPBytes(doc)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, finalBytes, 0644); err != nil {
		return err
	}
	return nil
}
