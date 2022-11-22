package gemfile

import (
	"bytes"
	"os"
	"regexp"

	"github.com/bearer/curio/pkg/detectors/dependencies/depsbase"
	"github.com/bearer/curio/pkg/util/file"
	"github.com/bearer/curio/pkg/util/linescanner"
	"github.com/rs/zerolog/log"
)

const (
	supportedFileFormatVersion = 1
	patternBundled             = "BUNDLED WITH"
	patternDependencies        = "DEPENDENCIES"
	patternPlatforms           = "PLATFORMS"
	patternRuby                = "RUBY VERSION"
	patternGit                 = "GIT"
	patternGem                 = "GEM"
	patternPath                = "PATH"
	patternPlugin              = "PLUGIN SOURCE"
	patternSpecs               = "  specs:"

	stateSource      = "source"
	stateDependency  = "dependency"
	statePlatform    = "platform"
	stateRuby        = "ruby"
	stateBundledWith = "bundled_with"
	stateUnknown     = "unknown"
)

var patternOptionsRegexp *regexp.Regexp
var patternOtherRegexp *regexp.Regexp
var patternNameAndVersionRegexp *regexp.Regexp

// SourceType is the name of a type
type SourceType struct {
	Name string
}

func Discover(file *file.FileInfo) (report *depsbase.DiscoveredDependency) {
	report = &depsbase.DiscoveredDependency{}
	report.Provider = "gemfile-lock"
	report.Language = "Ruby"
	report.PackageManager = "rubygems"

	fileBytes, err := os.ReadFile(file.AbsolutePath)
	if err != nil {
		log.Error().Msgf("%s: there was an error while opening the file: %s", report.Provider, err.Error())
		return nil
	}

	depfile := linescanner.New(bytes.NewBuffer(fileBytes))

	var state = stateUnknown
	var currentSourceType = SourceType{}
	for depfile.Scan() {
		line := depfile.Text()
		lineNumber := depfile.LineNumber()

		switch {

		case line == patternGit || line == patternGem || line == patternPath || line == patternPlugin:
			state = stateSource
			parseSource(report, &currentSourceType, line, lineNumber)

		case line == patternDependencies:
			state = stateDependency

		case line == patternPlatforms:
			state = statePlatform

		case line == patternRuby:
			state = stateRuby

		case line == patternBundled:
			state = stateBundledWith

		case patternOtherRegexp.MatchString(line):
			state = stateUnknown

		case state != stateUnknown:
			switch state {

			case stateSource:
				parseSource(report, &currentSourceType, line, lineNumber)

			case stateDependency:
				// TODO

			case statePlatform:
				// TODO

			case stateRuby:
				// TODO

			case stateBundledWith:
				// TODO
			}
		}
	}

	return report
}

func init() {
	// compile regexps
	patternOtherRegexp = regexp.MustCompile(`^[^\s]`)
	patternOptionsRegexp = regexp.MustCompile(`(?i)^  ([a-z]+): (.*)$`)
	patternNameAndVersionRegexp = regexp.MustCompile(`^( {2}| {4}| {6})([^\s]*?)(?: \(([^-]*)(?:-(.*))?\))?(\!)?$`)
}

func parseSource(report *depsbase.DiscoveredDependency, currentSourceType *SourceType, line string, lineNumber int) {
	switch {

	case line == patternSpecs:
		switch currentSourceType.Name {

		case patternPath:

		case patternGit:

		case patternGem:

		case patternPlugin:
		}

	case patternOptionsRegexp.MatchString(line):

	case line == patternGit || line == patternGem || line == patternPath || line == patternPlugin:
		currentSourceType.Name = line
	default:
		if currentSourceType.Name != patternGem {
			return
		}
		parseSpec(report, line, lineNumber)
	}
}

func parseSpec(document *depsbase.DiscoveredDependency, line string, lineNumber int) {
	var matches []string
	if matches = patternNameAndVersionRegexp.FindStringSubmatch(line); matches == nil {
		return
	}

	spaces := matches[1]
	name := matches[2] // pg
	version := matches[3]

	switch len(spaces) {

	case 4:
		document.Dependencies = append(document.Dependencies, depsbase.Dependency{Name: name, Line: int64(lineNumber), Column: 0, Version: version})

	case 6:
		// TODO: parse the dependencies of current spec
	}
}
