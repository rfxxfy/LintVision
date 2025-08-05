package extensions

var codeExts = map[string]struct{}{
	".go":   {},
	".py":   {},
	".js":   {},
	".java": {},
	".c":    {},
	".cpp":  {},
	".h":    {},
	".cs":   {},
	".rb":   {},
	".php":  {},
	".ts":   {},
}

func IsCodeExtension(ext string) bool {
	_, ok := codeExts[ext]
	return ok
}
