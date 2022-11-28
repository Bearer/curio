package schema

import (
	"github.com/bearer/curio/pkg/report/detectors"
	"github.com/bearer/curio/pkg/report/source"
)

const (
	SimpleTypeFunction = "function"
	SimpleTypeObject   = "object"
	SimpleTypeNumber   = "number"
	SimpleTypeDate     = "date"
	SimpleTypeString   = "string"
	SimpleTypeBool     = "boolean"
	SimpleTypeBinary   = "binary"
	SimpleTypeUnknown  = "unknown"
)

type Schema struct {
	ObjectName      string      `json:"object_name" yaml:"object_name"`
	ObjectUUID      string      `json:"-" yaml:"-"`
	FieldName       string      `json:"field_name" yaml:"field_name"`
	FieldUUID       string      `json:"-" yaml:"-"`
	FieldType       string      `json:"field_type" yaml:"field_type"`
	SimpleFieldType string      `json:"field_type_simple" yaml:"field_type_simple"`
	Classification  interface{} `json:"classification,omitempty" yaml:"classification,omitempty"`
	Parent          *Parent     `json:"parent,omitempty" yaml:"parent,omitempty"`
}

type Parent struct {
	// This is the starting line number, the very beginning of what's used by the custom detection
	LineNumber int    `json:"line_number,omitempty" yaml:"line_number,omitempty"`
	Content    string `json:"content,omitempty" yaml:"content,omitempty"`
}

type ReportSchema interface {
	AddSchema(detectorType detectors.Type, schema Schema, source source.Source)
}
