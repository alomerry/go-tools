package constant

var (
	MySQL = newDatabaseType("mysql")
)

var (
	Umami      = newDatabase(MySQL, "umami")
	WalineBlog = newDatabase(MySQL, "waline_blog")
)

type DatabaseType string

func newDatabaseType(name string) DatabaseType {
	return DatabaseType(name)
}

type Database struct {
	Type DatabaseType
	Name string
}

func newDatabase(dbType DatabaseType, name string) Database {
	return Database{dbType, name}
}
