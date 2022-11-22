package artifact

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"golang.org/x/xerrors"

	"github.com/bearer/curio/pkg/commands/process/balancer"
	"github.com/bearer/curio/pkg/commands/process/settings"
	"github.com/bearer/curio/pkg/commands/process/worker/work"
	"github.com/bearer/curio/pkg/flag"
	reportoutput "github.com/bearer/curio/pkg/report/output"
	outputhandler "github.com/bearer/curio/pkg/util/output"
	"github.com/bearer/curio/pkg/util/tmpfile"

	"github.com/bearer/curio/pkg/types"
)

// TargetKind represents what kind of artifact curio scans
type TargetKind string

const (
	TargetFilesystem TargetKind = "fs"
	TargetRepository TargetKind = "repo"
)

type ScannerConfig struct {
	Target   string
	Artifact types.Artifact
}

type Runner interface {
	// ScanFilesystem scans a filesystem
	ScanFilesystem(ctx context.Context, opts flag.Options) (types.Report, error)
	// ScanRepository scans repository
	ScanRepository(ctx context.Context, opts flag.Options) (types.Report, error)
	// Report a writes a report
	Report(scanSettings settings.Config, report types.Report) error
	// Close closes runner
	Close(ctx context.Context) error
}

type runner struct {
	balancer   *balancer.Monitor
	reportPath string
}

// NewRunner initializes Runner that provides scanning functionalities.
func NewRunner(ctx context.Context, scanSettings settings.Config) Runner {
	r := &runner{}

	r.balancer = balancer.New(scanSettings)
	r.reportPath = tmpfile.Create(os.TempDir(), ".jsonl")

	return r
}

// Close closes everything
func (r *runner) Close(ctx context.Context) error {
	return nil
}

func (r *runner) ScanFilesystem(ctx context.Context, opts flag.Options) (types.Report, error) {
	return r.scanFS(ctx, opts)
}

func (r *runner) scanFS(ctx context.Context, opts flag.Options) (types.Report, error) {

	return r.scanArtifact(ctx, opts)
}

func (r *runner) ScanRepository(ctx context.Context, opts flag.Options) (types.Report, error) {

	return r.scanArtifact(ctx, opts)
}

func (r *runner) scanArtifact(ctx context.Context, opts flag.Options) (types.Report, error) {
	task := r.balancer.ScheduleTask(work.ProcessRequest{
		Repository: work.Repository{
			Dir:               opts.Target,
			PreviousCommitSHA: "",
			CommitSHA:         "",
		},
		FilePath: r.reportPath,
	})
	result := <-task.Done

	if result.Error != nil {
		return types.Report{}, result.Error
	}

	log.Debug().Msgf("report location: %s", r.reportPath)

	return types.Report{
		Path: r.reportPath,
	}, nil
}

// Run performs artifact scanning
func Run(ctx context.Context, opts flag.Options, targetKind TargetKind) (err error) {
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	defer func() {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warn().Msg("Increase --timeout value")
		}
	}()

	scanSettings, err := settings.FromOptions(opts)
	if err != nil {
		return err
	}

	r := NewRunner(ctx, scanSettings)
	defer r.Close(ctx)

	var report types.Report
	switch targetKind {
	case TargetFilesystem:
		if report, err = r.ScanFilesystem(ctx, opts); err != nil {
			return xerrors.Errorf("filesystem scan error: %w", err)
		}
	}

	if err = r.Report(scanSettings, report); err != nil {
		return xerrors.Errorf("report error: %w", err)
	}

	return nil
}

func (r *runner) Report(config settings.Config, report types.Report) error {
	// if output is defined we want to write only to file
	logger := outputhandler.StdOutLogger()
	if config.Report.Output != "" {
		reportFile, err := os.Create(config.Report.Output)
		if err != nil {
			return fmt.Errorf("error creating output file %w", err)
		}
		logger = outputhandler.PlainLogger(reportFile)
	}

	switch config.Report.Format {
	case flag.FormatJSON:
		// TODO: support table flag
		err := reportoutput.ReportTable(report, logger, config)
		if err != nil {
			return fmt.Errorf("error generating report %w", err)
		}
		// err := reportoutput.ReportJSON(report, logger, config)
		// if err != nil {
		// 	return fmt.Errorf("error generating report %w", err)
		// }
	case flag.FormatYAML:
		err := reportoutput.ReportYAML(report, logger, config)
		if err != nil {
			return fmt.Errorf("error generating report %w", err)
		}
	}
	return nil
}
