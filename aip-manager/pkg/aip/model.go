// pkg/aip/model.go
package aip

import "encoding/xml"

// Document represents the root <DOCUMENT> element of the .aip file.
type Document struct {
	XMLName    xml.Name    `xml:"DOCUMENT"`
	Type       string      `xml:"Type,attr"`
	Version    string      `xml:"version,attr"`
	OtherAttrs []xml.Attr  `xml:",any,attr"`
	Components []Component `xml:"COMPONENT"`
}

// Component represents a <COMPONENT> element.
// We use a generic component struct that can hold any set of rows.
type Component struct {
	CID  string `xml:"cid,attr"`
	Rows []Row  `xml:"ROW"`
}

// Row represents a <ROW> element within a component.
// It contains fields for all attributes we might need to read or modify.
// `omitempty` ensures that if a field is empty, its attribute is not written to the file.
// `OtherAttrs` is a catch-all for any attributes we don't explicitly define,
// ensuring we don't lose data when writing the file back.
type Row struct {
	XMLName         xml.Name   `xml:"ROW"`
	Property        string     `xml:"Property,attr,omitempty"`
	Value           string     `xml:"Value,attr,omitempty"`
	File            string     `xml:"File,attr,omitempty"`
	SourcePath      string     `xml:"SourcePath,attr,omitempty"`
	Component       string     `xml:"Component_,attr,omitempty"`
	Registry        string     `xml:"Registry,attr,omitempty"`
	Root            string     `xml:"Root,attr,omitempty"`
	Key             string     `xml:"Key,attr,omitempty"`
	Name            string     `xml:"Name,attr,omitempty"`
	BuildKey        string     `xml:"BuildKey,attr,omitempty"`
	PackageFileName string     `xml:"PackageFileName,attr,omitempty"`
	PackageFolder   string     `xml:"PackageFolder,attr,omitempty"` // Added
	OtherAttrs      []xml.Attr `xml:",any,attr"`
}
