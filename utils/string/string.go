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
