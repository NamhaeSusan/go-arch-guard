package presets

type Preset string

const (
	PresetDDD             Preset = "ddd"
	PresetCleanArch       Preset = "cleanarch"
	PresetLayered         Preset = "layered"
	PresetHexagonal       Preset = "hexagonal"
	PresetModularMonolith Preset = "modular_monolith"
	PresetConsumerWorker  Preset = "consumer_worker"
	PresetBatch           Preset = "batch"
	PresetEventPipeline   Preset = "event_pipeline"
)

func defaultBannedPkgNames() []string {
	return []string{"util", "common", "misc", "helper", "shared", "services"}
}

func defaultLegacyPkgNames() []string {
	return []string{"router", "bootstrap"}
}

func domainTopLevel() map[string]bool {
	return map[string]bool{
		"domain":        true,
		"orchestration": true,
		"pkg":           true,
	}
}

func domainLayout() (string, string, string) {
	return "domain", "orchestration", "pkg"
}
