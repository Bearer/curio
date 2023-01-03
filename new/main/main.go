package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v2"

	"github.com/bearer/curio/new/detector/evaluator"
	"github.com/bearer/curio/new/detector/implementation/custom"
	"github.com/bearer/curio/new/detector/implementation/generic/datatype"
	"github.com/bearer/curio/new/detector/implementation/ruby/object"
	"github.com/bearer/curio/new/detector/implementation/ruby/property"
	detectorset "github.com/bearer/curio/new/detector/set"
	detectortypes "github.com/bearer/curio/new/detector/types"
	"github.com/bearer/curio/new/language"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	lang, err := language.Get("ruby")
	if err != nil {
		return fmt.Errorf("failed to lookup language: %s", err)
	}

	propertyDetector, err := property.New(lang)
	if err != nil {
		return fmt.Errorf("failed to create property detector: %s", err)
	}
	defer propertyDetector.Close()

	objectDetector, err := object.New(lang)
	if err != nil {
		return fmt.Errorf("failed to create object detector: %s", err)
	}
	defer objectDetector.Close()

	datatypeDetector, err := datatype.New(lang)
	if err != nil {
		return fmt.Errorf("failed to create datatype detector: %s", err)
	}
	defer datatypeDetector.Close()

	rubyFileDetector, err := custom.New(
		lang,
		"ruby_file_detection",
		custom.Rule{
			Pattern: `
				Sentry.init do |$<CONFIG:identifier>|
					$<CONFIG>.before_breadcrumb = lambda do |$<BREADCRUMB:identifier>, hint|
						$<BREADCRUMB>.message = $<MESSAGE>
					end
				end`,
			Filters: []custom.Filter{
				{
					Variable:  "MESSAGE",
					Detection: "datatype",
				},
			},
		},
	)
	// rubyFileDetector, err := custom.New(
	// 	lang,
	// 	"ruby_file_detection",
	// 	custom.Rule{
	// 		Pattern: `
	// 			$<LIBRARY:constant>.open do
	// 				$<BLOCK_EXPRESSION:_>
	// 			end`,
	// 		Filters: []custom.Filter{
	// 			{
	// 				Variable: "LIBRARY",
	// 				Values:   []string{"CSV", "File"},
	// 			},
	// 			{
	// 				Variable:         "BLOCK_EXPRESSION",
	// 				ContainsDataType: true,
	// 			},
	// 		},
	// 	},
	// )
	if err != nil {
		return fmt.Errorf("failed to create ruby file detector: %s", err)
	}
	defer rubyFileDetector.Close()

	detectorSet, err := detectorset.New([]detectortypes.Detector{
		propertyDetector,
		objectDetector,
		datatypeDetector,
		rubyFileDetector,
	})
	if err != nil {
		return fmt.Errorf("failed to create detector set: %s", err)
	}

	tree, err := lang.Parse(`

Sentry.init do |config|
	config.before_breadcrumb = lambda do |breadcrumb, hint|
		breadcrumb.message = { user: { first_name: "" } }     # 5
		bf.message = { user: { first_name: "" } }             # 6
	end

	other.before_breadcrumb = lambda do |breadcrumb, hint|
		breadcrumb.message = { user: { first_name: "" } }     # 10
	end
end

FSentry.init do |config|
	config.before_breadcrumb = lambda do |breadcrumb, hint|
		breadcrumb.message = { user: { first_name: "" } }     # 5
	end
end

	`)
	// 	tree, err := lang.Parse(`

	// class Abc
	// 	def f
	// 		CSV.open("path/to/user.csv", "wb") do |csv|																																	 # 5
	// 			users.each do |user1|
	// 				user = { first_name: "" }                                                                                # 7
	// 				csv << user.values
	// 			end
	// 		end

	// 		Other.open("path/to/user.csv", "wb") do |csv|
	// 			users.each do |user1|
	// 				user = { first_name: "" }																															                   # 14
	// 				csv << user.values
	// 			end
	// 		end
	// 	end
	// end

	// 	`)
	// 	tree, err := lang.Parse(`

	// class Abc
	// 	def f
	// 		user = { first_name: "hello", last_name: "there" }   # 5
	// 		HTTP.post("http://api.com", user)                    # 6

	// 		user = { first_name: "hello2", last_name: "there2" } # 8
	// 		HTTP.post("http://api.com", user)                    # 9

	// 		HTTP.post("http://api.com", { x: user })             # 11

	// 		x = { first_name: "hello3", last_name: "there3" }    # 13
	// 		HTTP.post("http://api.com", { user: x })             # 14
	// 	end
	// end

	// 	`)
	if err != nil {
		return err
	}
	defer tree.Close()

	evaluator := evaluator.New(lang, detectorSet, tree)

	detections, err := evaluator.ForTree(tree.RootNode(), "ruby_file_detection")
	if err != nil {
		return fmt.Errorf("failed to detect: %s", err)
	}

	detectionsYAML, _ := yaml.Marshal(report(detections))
	log.Printf("GOT:\n%s\n", detectionsYAML)

	return nil
}

type Location struct {
	Content    string
	LineNumber int
}
type ReportDetection struct {
	MatchLocation   Location
	ContextLocation *Location
	Data            interface{}
}

func report(detections []*detectortypes.Detection) []ReportDetection {
	result := make([]ReportDetection, len(detections))

	for i, detection := range detections {
		var contextLocation *Location

		if detection.ContextNode != nil {
			contextLocation = &Location{
				Content:    detection.ContextNode.Content(),
				LineNumber: detection.ContextNode.LineNumber(),
			}
		}

		result[i] = ReportDetection{
			MatchLocation: Location{
				Content:    detection.MatchNode.Content(),
				LineNumber: detection.MatchNode.LineNumber(),
			},
			Data:            detection.Data,
			ContextLocation: contextLocation,
		}
	}

	return result
}
