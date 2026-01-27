// pkg/enigma/profile.go
// ----------------------------------------------------------------------------
// Enigma Protector profile (.enigma64) path rewriter.
//
// Problem
//   Profiles often contain absolute paths for inputs/outputs. In CI workspaces,
//   absolute paths are ephemeral, so profiles must be adjusted at runtime.
//
// Approach
//   Rewrite only the *path-bearing* nodes to point into the job's actual
//   Unpacked/Packed directories (keeping file basenames). Everything else is
//   preserved as-is via a streaming transform.
//
// Rewritten nodes
//   1) Top-level:
//        <Input><FileName>...</FileName></Input>        => Unpacked\<basename>
//        <Output><FileName>...</FileName></Output>      => Packed\<basename>
//   2) Advanced list under <AdvanceInput>:
//        <Files><File><Input>...</Input>                => Unpacked\<basename>
//        <Files><File><Output>...</Output>              => Packed\<basename>
//
// Output
//   Rewritten profile is written to: build/<yyyy-MM-dd>/enigma_profile.enigma64
//   The absolute path is returned for convenience.
//
// XML notes
//   Some profiles include an XML declaration `<?xml ...?>`. The Go encoder will
//   only accept that declaration as the *very first* encoded token. Real files
//   can have BOM/whitespace before it; to avoid encoder errors, we simply
//   skip re-emitting the `xml` processing instruction. Enigma accepts profiles
//   without the declaration.
// ----------------------------------------------------------------------------

package enigma

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateProfileWithMappedPaths reads profilePath, rewrites path-bearing nodes
// to point into unpackedDir/packedDir, writes the result into
// build/<yyyy-MM-dd>/enigma_profile.enigma64, and returns that absolute path.
func GenerateProfileWithMappedPaths(profilePath, unpackedDir, packedDir string) (string, error) {
	raw, err := os.ReadFile(profilePath)
	if err != nil {
		return "", fmt.Errorf("read profile: %w", err)
	}
	mod, err := rewriteEnigmaProfileXML(raw, unpackedDir, packedDir)
	if err != nil {
		return "", err
	}
	outDir := filepath.Join("build", time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("create out dir: %w", err)
	}
	outPath := filepath.Join(outDir, "enigma_profile.enigma64")
	if err := os.WriteFile(outPath, mod, 0o644); err != nil {
		return "", fmt.Errorf("write rewritten profile: %w", err)
	}
	if a, err := filepath.Abs(outPath); err == nil {
		return a, nil
	}
	return outPath, nil
}

// rewriteEnigmaProfileXML performs a streaming transform, changing only the
// intended CharData nodes while preserving all other tags/attributes.
//
// IMPORTANT: We do *not* re-emit the XML declaration (`<?xml ...?>`) because
// the encoder only allows it as the very first token, and leading BOM/whitespace
// can violate that constraint. Enigma does not require the declaration.
func rewriteEnigmaProfileXML(src []byte, unpackedDir, packedDir string) ([]byte, error) {
	dec := xml.NewDecoder(bytes.NewReader(src))
	var out bytes.Buffer
	enc := xml.NewEncoder(&out)

	var stack []string
	writeTok := func(tok xml.Token) error { return enc.EncodeToken(tok) }

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xml decode: %w", err)
		}

		switch t := tok.(type) {
		case xml.ProcInst:
			// Skip XML declaration; emit other processing instructions unchanged.
			if strings.EqualFold(t.Target, "xml") {
				// do not write; move on
				continue
			}
			if err := writeTok(t); err != nil {
				return nil, err
			}

		case xml.StartElement:
			stack = append(stack, t.Name.Local)
			if err := writeTok(t); err != nil {
				return nil, err
			}

		case xml.EndElement:
			if err := writeTok(t); err != nil {
				return nil, err
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			newText := []byte(t)
			trim := strings.TrimSpace(string(t))
			if trim != "" {
				inAdv := contains(stack, "AdvanceInput")
				switch {
				// Top-level <Input><FileName> (outside AdvanceInput)
				case !inAdv && hasSuffix(stack, "Input", "FileName"):
					newText = []byte(joinUnpacked(unpackedDir, trim))
				// Top-level <Output><FileName> (outside AdvanceInput)
				case !inAdv && hasSuffix(stack, "Output", "FileName"):
					newText = []byte(joinPacked(packedDir, trim))
				// Advanced list entries under <AdvanceInput>…<Files>…<File>…
				case inAdv && hasSuffix(stack, "File", "Input"):
					newText = []byte(joinUnpacked(unpackedDir, trim))
				case inAdv && hasSuffix(stack, "File", "Output"):
					newText = []byte(joinPacked(packedDir, trim))
				}
			}
			if err := writeTok(xml.CharData(newText)); err != nil {
				return nil, err
			}

		default:
			// Comments, directives, etc.
			if err := writeTok(t); err != nil {
				return nil, err
			}
		}
	}

	if err := enc.Flush(); err != nil {
		return nil, fmt.Errorf("xml encode: %w", err)
	}
	return out.Bytes(), nil
}

// hasSuffix reports whether the element stack ends with the given path.
func hasSuffix(stack []string, elems ...string) bool {
	if len(stack) < len(elems) {
		return false
	}
	start := len(stack) - len(elems)
	for i := range elems {
		if stack[start+i] != elems[i] {
			return false
		}
	}
	return true
}

// contains reports whether the stack contains a name anywhere.
func contains(stack []string, name string) bool {
	for _, s := range stack {
		if s == name {
			return true
		}
	}
	return false
}

// joinUnpacked maps an original absolute/relative path to Unpacked\<basename>.
func joinUnpacked(unpackedDir, original string) string {
	base := filepath.Base(strings.TrimSpace(original))
	return filepath.Join(unpackedDir, base)
}

// joinPacked maps an original absolute/relative path to Packed\<basename>.
func joinPacked(packedDir, original string) string {
	base := filepath.Base(strings.TrimSpace(original))
	return filepath.Join(packedDir, base)
}
