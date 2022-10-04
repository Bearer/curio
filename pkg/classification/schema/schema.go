package schema

import (
	"github.com/bearer/curio/pkg/parser/datatype"
)

type Classifier struct {
	config Config
}

type Config struct {
}

func New(config Config) *Classifier {
	return &Classifier{}
}

func (classifier *Classifier) Classify(data datatype.DataType) (ClassifiedDatatype, error) {
	// todo: implement interface classification (bigbear etc...)
	return ClassifiedDatatype{}, nil
}
