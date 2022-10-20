package worker

import (
	"encoding/json"

	"net/http"
	"runtime"

	classsification "github.com/bearer/curio/pkg/classification"
	config "github.com/bearer/curio/pkg/commands/process/settings"
	"github.com/bearer/curio/pkg/commands/process/worker/blamer"
	"github.com/bearer/curio/pkg/commands/process/worker/work"
	"github.com/bearer/curio/pkg/detectors"
	"github.com/bearer/curio/pkg/scanner"
)

func Start(port string, config config.Config) error {
	classifier := classsification.NewClassifier(&classsification.Config{Config: config})
	err := detectors.SetupCustomDetector(config.CustomDetector.RulesConfig)
	if err != nil {
		return err
	}

	err = http.ListenAndServe(`localhost`+port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close() //golint:all,errcheck

		switch r.URL.Path {
		case work.RouteStatus:
			var scanRequest work.ProcessRequest
			json.NewDecoder(r.Body).Decode(&scanRequest) //nolint:all,errcheck

			response := &work.ProcessResponse{}

			if scanRequest.CustomDetectorConfig != nil {
				err := detectors.SetupCustomDetector(scanRequest.CustomDetectorConfig)
				if err != nil {
					response.Error = err
				}
			}

			json.NewEncoder(rw).Encode(response) //nolint:all,errcheck
		case work.RouteProcess:
			runtime.GC()

			var scanRequest work.ProcessRequest
			json.NewDecoder(r.Body).Decode(&scanRequest) //nolint:all,errcheck

			blamer := blamer.New(scanRequest.Dir, scanRequest.BlameRevisionsFilePath, scanRequest.PreviousCommitSHA)

			response := &work.ProcessResponse{}
			var filesList []string
			for _, file := range scanRequest.Files {
				filesList = append(filesList, file.FilePath)
			}

			response.Error = scanner.Scan(scanRequest.Dir, filesList, blamer, scanRequest.FilePath, classifier)

			json.NewEncoder(rw).Encode(response) //nolint:all,errcheck
		default:
			rw.WriteHeader(http.StatusNotFound)
		}
	}))

	return err
}
