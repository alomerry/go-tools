package pusher

type Pusher interface {
	Init() error
	Push(filePath string, remotePath string) error
}
