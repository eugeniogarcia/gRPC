package match

import (
	"reflect"

	"github.com/golang/mock/gomock"
)

//Definimos un tipo que implementa el interface de un matcher
type deTipo struct{ t string }

func (o *deTipo) Matches(x interface{}) bool {
	return reflect.TypeOf(x).String() == o.t
}

func (o *deTipo) String() string {
	return "is of type " + o.t
}

//EsTipo nuestro propio matcher
func EsTipo(t string) gomock.Matcher {
	return &deTipo{t}
}
