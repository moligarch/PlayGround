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

// CommonFileIndex holds a two-tiered map for common files to resolve conflicts.
type CommonFileIndex struct {
	bySubdirAndFileName map[string]string
	byFileName          map[string]string
}

// FileIndex holds separate maps for each category of source files.
type FileIndex struct {
	CommonFiles   *CommonFileIndex
	PackedFiles   map[string]string
	UnpackedFiles map[string]string
}

// buildFileIndex walks the search directories and creates a detailed index of all files.
func buildFileIndex(baseDir string) (*FileIndex, error) {
	index := &FileIndex{
		CommonFiles: &CommonFileIndex{
			bySubdirAndFileName: make(map[string]string),
			byFileName:          make(map[string]string),
		},
		PackedFiles:   make(map[string]string),
		UnpackedFiles: make(map[string]string),
	}

	// --- Index Common Files ---
	commonDir := filepath.Join(baseDir, "Common")
	if _, err := os.Stat(commonDir); !os.IsNotExist(err) {
		err := filepath.WalkDir(commonDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				fileName := d.Name()
				parentDirName := filepath.Base(filepath.Dir(path))

				// Store by filename as a fallback (e.g., "sqlite3.dll")
				index.CommonFiles.byFileName[fileName] = path

				// Store by specific key if it's in a subdirectory (e.g., "PhoenixAM/sqlite3.dll")
				if parentDirName != "Common" {
					specificKey := filepath.ToSlash(filepath.Join(parentDirName, fileName))
					index.CommonFiles.bySubdirAndFileName[specificKey] = path
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// --- Index Packed and Unpacked Files ---
	for _, dirName := range []string{"Packed", "Unpacked"} {
		targetMap := index.PackedFiles
		if dirName == "Unpacked" {
			targetMap = index.UnpackedFiles
		}
		dir := filepath.Join(baseDir, dirName)
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() {
					targetMap[d.Name()] = path
				}
				return nil
			})
		}
	}

	return index, nil
}

// ResolveSourcePaths intelligently finds files by name and replaces their SourcePath attributes
// with the correct absolute path based on the specified buildType and subdirectory.
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

				var absolutePath string
				var found bool

				originalPath := filepath.ToSlash(row.SourcePath)
				parts := strings.Split(originalPath, "/")
				fileName := parts[len(parts)-1]

				// --- NEW PRIORITY-BASED LOOKUP LOGIC ---
				// 1. Prioritize a specific subdirectory lookup in Common files.
				if len(parts) > 2 {
					parentDirName := parts[len(parts)-2]
					if parentDirName != "Common" && parentDirName != "Packed" && parentDirName != "Unpacked" {
						lookupKey := filepath.ToSlash(filepath.Join(parentDirName, fileName))
						absolutePath, found = index.CommonFiles.bySubdirAndFileName[lookupKey]
					}
				}

				// 2. If not found, check Packed or Unpacked based on buildType.
				if !found {
					if buildType == "packed" {
						absolutePath, found = index.PackedFiles[fileName]
					} else {
						absolutePath, found = index.UnpackedFiles[fileName]
					}
				}

				// 3. If still not found, fall back to a generic search in Common files.
				if !found {
					absolutePath, found = index.CommonFiles.byFileName[fileName]
				}
				// --- END OF NEW LOGIC ---

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
