package user_test

import (
	"errors"
	"testing"
	"testing/mocks"
	"testing/user"

	"github.com/golang/mock/gomock"
)

func TestUse(t *testing.T) {
	//Controlador del test
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	//Crea una instancia del mock
	mockDoer := mocks.NewMockDoer(mockCtrl)
	//Crea una instancia del tipo a probar, usando el mock de Doer, en lugar de Doer
	testUser := &user.User{Doer: mockDoer}

	//Especifica las assertions
	//Esperamos que el método DoSomething se haya llamado con los argumentos especificados, y el tipo de retorn.
	mockDoer.EXPECT().DoSomething(123, "Hello GoMock").Return(nil).Times(1)

	testUser.Use()
}

func TestUseReturnsErrorFromDo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	dummyError := errors.New("dummy error")

	mockDoer := mocks.NewMockDoer(mockCtrl)
	testUser := &user.User{Doer: mockDoer}

	mockDoer.EXPECT().DoSomething(123, "Hello GoMock").Return(dummyError).Times(1)

	err := testUser.Use()

	if err != dummyError {
		t.Fail()
	}
}
