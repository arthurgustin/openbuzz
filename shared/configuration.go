package shared

type AppConfig struct {
	PgPort     int    `split_words:"true" default:"5432"`
	Port       int    `split_words:"true" default:"1346"`
	PgHost     string `split_words:"true" default:"localhost"`
	PgUser     string `split_words:"true" default:"postgres"`
	PgPassword string `split_words:"true" default:"postgres"`
	PgDbName   string `split_words:"true" default:"openbuzz"`
}
