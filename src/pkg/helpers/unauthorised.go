package helpers

type UnauthorisedError struct {
	Err string
}

func (u *UnauthorisedError) Error() string {
	return u.Err
}
