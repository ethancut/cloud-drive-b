package database

type DatabaseError struct {
	Message string
}

func (e *DatabaseError) Error() string {
	return e.Message
}

var InvalidPasswordError = &DatabaseError{Message: "Invalid password"}
var AccountNotFoundError = &DatabaseError{Message: "Account not found"}
