package account

type ErrAddressExists struct{}

func (e *ErrAddressExists) Error() string {
	return "address already exists"
}

type ErrAddressNotFound struct{}

func (e *ErrAddressNotFound) Error() string {
	return "address not found"
}
