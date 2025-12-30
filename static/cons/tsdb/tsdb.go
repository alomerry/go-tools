package tsdb

type TagOp string

const (
	OpEqual TagOp = "="
)

func ToTagOp(op string) TagOp {
	switch op {
	case "=":
		return OpEqual
	default:
		return OpEqual
	}
}
