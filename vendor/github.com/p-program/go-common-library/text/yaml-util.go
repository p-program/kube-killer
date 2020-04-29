package text

func RemoveYAMLcomment(oldText string) (resultText string) {
	array := strings.Split(oldText, LineBreak)
	// https://medium.com/@thuc/8-notes-about-strings-builder-in-golang-65260daae6e9
	sb := &strings.Builder{}
	for _, singleLine := range array {
		if strings.HasPrefix(singleLine, "#") || strings.HasPrefix(singleLine, "//") {
			continue
		}
		sb.WriteString(singleLine + LineBreak)
	}
	resultText = sb.String()
	return resultText
}
