package flag

var (
	SkipPathFlag = Flag{
		Name:       "skip-path",
		ConfigName: "scan.skip-path",
		Value:      []string{},
		Usage:      "specify the comma separated files and directories to skip (supports * syntax), eg. --skip-path users/*.go,users/admin.sql",
	}
	DebugFlag = Flag{
		Name:       "debug",
		ConfigName: "scan.debug",
		Value:      false,
		Usage:      "enable debug logs",
	}
)

type ScanFlagGroup struct {
	SkipPathFlag *Flag
	DebugFlag    *Flag
}

type ScanOptions struct {
	Target   string   `json:"target"`
	SkipPath []string `json:"skip_path"`
	Debug    bool     `json:"debug"`
}

func NewScanFlagGroup() *ScanFlagGroup {
	return &ScanFlagGroup{
		SkipPathFlag: &SkipPathFlag,
		DebugFlag:    &DebugFlag,
	}
}

func (f *ScanFlagGroup) Name() string {
	return "Scan"
}

func (f *ScanFlagGroup) Flags() []*Flag {
	return []*Flag{
		f.SkipPathFlag,
		f.DebugFlag,
	}
}

func (f *ScanFlagGroup) ToOptions(args []string) (ScanOptions, error) {
	var target string
	if len(args) == 1 {
		target = args[0]
	}

	return ScanOptions{
		SkipPath: getStringSlice(f.SkipPathFlag),
		Debug:    getBool(f.DebugFlag),
		Target:   target,
	}, nil
}
