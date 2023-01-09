package scanner

import (
	"fmt"

	"github.com/bearer/curio/new/detector/composition/ruby"
	"github.com/bearer/curio/new/detector/types"
	"github.com/bearer/curio/pkg/commands/process/settings"
	"github.com/bearer/curio/pkg/report"
	"github.com/bearer/curio/pkg/util/file"
	"github.com/rs/zerolog/log"
)

type language struct {
	name        string
	composition types.Composition
}

type scannerType []language

var scanner scannerType

func (scanner scannerType) Close() {
	for _, language := range scanner {
		language.composition.Close()
	}
}

func Setup(config map[string]settings.Rule) (err error) {
	var toInstantiate = []struct {
		constructor func(map[string]settings.Rule) (types.Composition, error)
		name        string
	}{{
		constructor: ruby.New,
		name:        "ruby",
	}}

	for _, instantiatior := range toInstantiate {
		composition, err := instantiatior.constructor(config)
		if err != nil {
			return fmt.Errorf("failed to instantiate composition %s:%s", instantiatior.name, err)
		}

		scanner = append(scanner, language{
			name:        instantiatior.name,
			composition: composition,
		})
	}

	return err
}

func Detect(report report.Report, file *file.FileInfo) (err error) {
	log.Debug().Msg("calling new detect")

	for _, language := range scanner {
		if err := language.composition.DetectFromFile(report, file); err != nil {
			return fmt.Errorf("%s failed to detect in file %s: %s", language.name, file.AbsolutePath, err)
		}
	}

	return nil
}
