package access

import "errors"

var ErrServicesUsersAccessDenied = errors.New("access denied")

const (
	zero      = "0"
	read      = "1"
	write     = "2"
	readWrite = "3"
	admin     = "4"
)

func IsZeorAccess(access string) bool {
	return validAccess(access) && access == zero
}

func validAccess(access string) bool {
	if n := len(access); n == 0 || n > 1 || access > admin {
		return false
	}
	return true
}

func WriteAccess(access string) bool {
	return validAccess(access) && access >= write
}

func ReadAccess(access string) bool {
	return validAccess(access) && (access == read || access >= readWrite)

}

func AdminAccess(access string) bool {
	return validAccess(access) && access == admin
}
