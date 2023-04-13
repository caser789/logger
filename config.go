package log

import "github.com/caser789/logger/internal/utils/env"

var (
	defaultConfig = &Config{
		Level:      InfoLvl,
		SplitLevel: SplitNone,
		PrintToStd: PrintToStd_NONE,
	}
	splitLogConfig = &Config{
		Level:      InfoLvl,
		SplitLevel: SplitDebug,
		PrintToStd: PrintToStd_NONE,
	}
)

func getDefaultConfig() *Config {
	if env.IsSplitLog() {
		return splitLogConfig
	}
	return defaultConfig
}
