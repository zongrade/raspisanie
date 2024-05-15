package myerror

type MyError struct {
	ErrSring string
}

func (mE MyError) Error() string {
	return mE.ErrSring
}

func CreateError(eS string) MyError {
	return MyError{
		ErrSring: eS,
	}
}
