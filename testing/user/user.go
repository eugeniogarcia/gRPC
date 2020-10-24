package user

import "testing/doer"

//User tipo que usa el interface que queremos mockear
type User struct {
	Doer doer.Doer
}

//Use metodo que hace uso del interface que queremos mockear
func (u *User) Use() error {
	return u.Doer.DoSomething(123, "Hello GoMock")
}
