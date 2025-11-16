package yaml

import (
	"bytes"

	"github.com/f0resee/stdlib/apollo/utils"
	"github.com/spf13/viper"
)

var vp = viper.New()

func init() {
	vp.SetConfigType("yaml")
}

type Parser struct {
}

func (d *Parser) Parse(configContent interface{}) (map[string]interface{}, error) {
	content, ok := configContent.(string)
	if !ok {
		return nil, nil
	}
	if utils.Empty == content {
		return nil, nil
	}

	buffer := bytes.NewBufferString(content)
	err := vp.ReadConfig(buffer)
	if err != nil {
		return nil, err
	}

	return convertToMap(vp), nil
}

func convertToMap(vp *viper.Viper) map[string]interface{} {
	if vp == nil {
		return nil
	}

	m := make(map[string]interface{})
	for _, key := range vp.AllKeys() {
		m[key] = vp.Get(key)
	}
	return m
}
