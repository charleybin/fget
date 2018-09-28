package utils

var globalDebug = true

func IsDebug() bool {
	return globalDebug
}

func SetDebug(dbg bool) {
	globalDebug = dbg
}
