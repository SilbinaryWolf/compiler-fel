package util

func IsSelfClosingTagName(name string) bool {
	// Source: https://www.quora.com/Which-HTML-tags-are-self-closing
	return name == "br" || name == "area" ||
		name == "base" || name == "col" || name == "command" ||
		name == "embed" || name == "hr" || name == "img" ||
		name == "input" || name == "keygen" || name == "link" ||
		name == "meta" || name == "param" || name == "source" ||
		name == "track" || name == "wbr"
}
