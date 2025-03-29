package source

// SQLScan - для использования с generic во время сканирования объектов
type SQLScan interface {
	Scan(dest ...any) error
}
