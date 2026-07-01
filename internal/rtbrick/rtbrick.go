package rtbrick

type PPDPort struct {
	Name   string `json:"name"`
	I2CBus int    `json:"i2c_bus"`
}
type PPDConfiguration struct {
	Ports []PPDPort `json:"port"`
}

type ImageMetadata struct {
	Manufacturer string `yaml:"Manufacturer"`
	Model        string `yaml:"Model"`
}
