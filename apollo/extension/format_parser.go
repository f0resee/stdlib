package extension

import (
	"github.com/f0resee/stdlib/apollo/constant"
	"github.com/f0resee/stdlib/apollo/utils/parse"
)

var formatParser = make(map[constant.ConfigFileFormat]parse.ContentParser, 0)

func AddFormatParser(key constant.ConfigFileFormat, contentParser parse.ContentParser) {
	formatParser[key] = contentParser
}

func GetFormatParser(key constant.ConfigFileFormat) parse.ContentParser {
	return formatParser[key]
}
