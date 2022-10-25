package datatype

import (
	"sort"
	"strings"

	"github.com/bearer/curio/pkg/parser"
	"github.com/bearer/curio/pkg/parser/nodeid"
	"github.com/bearer/curio/pkg/report/detections"
	"github.com/bearer/curio/pkg/report/detectors"
	"github.com/bearer/curio/pkg/report/schema"
)

type DataType struct {
	Node       *parser.Node
	Name       string
	Type       string
	TextType   string
	Properties map[string]*DataType
	IsHelper   bool // helper dataTypes and their child datatypes don't get exported
	UUID       string
}

func (datatype *DataType) GetName() string {
	return datatype.Name
}

func (datatype *DataType) GetNode() *parser.Node {
	return datatype.Node
}

func (datatype *DataType) GetProperties() map[string]*DataType {
	return datatype.Properties
}

func (datatype *DataType) GetUUID() string {
	return datatype.UUID
}

func (datatype *DataType) GetIsHelper() bool {
	return datatype.IsHelper
}

func (datatype *DataType) GetTextType() string {
	return datatype.TextType
}

func (datatype *DataType) GetType() string {
	return datatype.Type
}

type DataTypable interface {
	GetUUID() string
	GetIsHelper() bool
	GetTextType() string
	GetType() string
	GetName() string
	GetNode() *parser.Node
	GetProperties() map[string]DataTypable
}

func ExportSchemas[D DataTypable](report detections.ReportDetection, detectionType detections.DetectionType, detectorType detectors.Type, idGenerator nodeid.Generator, ignoreFirst bool, values map[parser.NodeID]D) {
	sortedDataTypes := SortParserMap(values)

	parentName := ""
	parentUUID := idGenerator.GenerateId()
	for _, value := range sortedDataTypes {
		dataTypeToSchema(report, detectionType, detectorType, idGenerator, value, parentName, parentUUID, !ignoreFirst)
	}
}

func dataTypeToSchema[D DataTypable](report detections.ReportDetection, detectionType detections.DetectionType, detectorType detectors.Type, idGenerator nodeid.Generator, dataType D, parentName string, parentUUID string, shouldExport bool) {
	if dataType.GetIsHelper() {
		return
	}

	selfUUID := dataType.GetUUID()
	if dataType.GetUUID() == "" {
		selfUUID = idGenerator.GenerateId()
	}

	selfName := dataType.GetName()

	if shouldExport {
		report.AddDetection(detectionType, detectorType, dataType.GetNode().Source(false), schema.Schema{
			ObjectName:      parentName,
			FieldName:       selfName,
			ObjectUUID:      parentUUID,
			FieldUUID:       selfUUID,
			FieldType:       dataType.GetTextType(),
			SimpleFieldType: dataType.GetType(),
		})
	}

	sortedProperties := SortStringMap(dataType.GetProperties())

	for _, property := range sortedProperties {
		dataTypeToSchema(report, detectionType, detectorType, idGenerator, property, selfName, selfUUID, true)
	}
}

func SortStringMap[D DataTypable](input map[string]D) []D {
	var sortedDataTypes []D
	for _, value := range input {
		sortedDataTypes = append(sortedDataTypes, value)
	}
	return SortSlice(sortedDataTypes)
}

func SortParserMap[D DataTypable](input map[parser.NodeID]D) []D {
	var sortedDataTypes []D
	for _, value := range input {
		sortedDataTypes = append(sortedDataTypes, value)
	}
	return SortSlice(sortedDataTypes)
}

func SortSlice[D DataTypable](input []D) []D {
	sort.Slice(input, func(i, j int) bool {
		lineNumberA := input[i].GetNode().Source(false).LineNumber
		lineNumberB := input[j].GetNode().Source(false).LineNumber

		if *lineNumberA != *lineNumberB {
			return *lineNumberA < *lineNumberB
		}

		columnNumberA := input[i].GetNode().Source(false).ColumnNumber
		columnNumberB := input[j].GetNode().Source(false).ColumnNumber

		if *columnNumberA != *columnNumberB {
			return *columnNumberA < *columnNumberB
		}

		if strings.Compare(input[i].GetName(), input[j].GetName()) == -1 {
			return true
		}

		return false
	})
	return input
}
