package string

const (
	Space   = " "
	Tab     = "\t"
	NewLine = "\n"
)

func FirstNotBlank(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return ""
}

func Limit(str string, limit int) string {
	if len(str) > limit {
		return str[:limit]
	}
	return str
}
