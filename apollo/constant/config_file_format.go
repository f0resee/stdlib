package constant

type ConfigFileFormat string

const (
	Properties ConfigFileFormat = ".properties"
	XML        ConfigFileFormat = ".xml"
	JSON       ConfigFileFormat = ".json"
	YML        ConfigFileFormat = ".yml"
	YAML       ConfigFileFormat = ".yaml"
	DEFAULT    ConfigFileFormat = ""
)
