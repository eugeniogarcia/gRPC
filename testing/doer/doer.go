package doer

//go:generate mockgen -source .\doer.go -destination ..\mocks\mock_doer.go -package mocks

//Doer interface a mockear
type Doer interface {
	DoSomething(int, string) error
}
