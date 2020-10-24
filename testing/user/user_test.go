package user_test

import (
	"errors"
	"testing"
	miMatcher "testing/match"
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

func TestCustomMatcher(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDoer := mocks.NewMockDoer(mockCtrl)
	testUser := &user.User{Doer: mockDoer}

	mockDoer.EXPECT().DoSomething(123, miMatcher.EsTipo("string")).Return(nil).Times(1)

	testUser.Use()
}

func TestOrden(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDoer := mocks.NewMockDoer(mockCtrl)
	testUser := &user.User{Doer: mockDoer}

	llamaPrimero := mockDoer.EXPECT().DoSomething(0, "first this")
	llamaA := mockDoer.EXPECT().DoSomething(1, "then this").After(llamaPrimero)
	_ = mockDoer.EXPECT().DoSomething(2, "or this").After(llamaPrimero)
	_ = mockDoer.EXPECT().DoSomething(3, "finally").After(llamaA)

	testUser.UseSeveral("first this", "then this", "or this", "finally")
}

func TestOrdenBis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDoer := mocks.NewMockDoer(mockCtrl)
	testUser := &user.User{Doer: mockDoer}

	gomock.InOrder(
		mockDoer.EXPECT().DoSomething(0, "first this"),
		mockDoer.EXPECT().DoSomething(1, "then this"),
		mockDoer.EXPECT().DoSomething(2, "or this"),
		mockDoer.EXPECT().DoSomething(3, "finally"),
	)

	testUser.UseSeveral("first this", "then this", "or this", "finally")
}

func TestAcciones(t *testing.T) {
	//Controlador del test
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	//Crea una instancia del mock
	mockDoer := mocks.NewMockDoer(mockCtrl)
	//Crea una instancia del tipo a probar, usando el mock de Doer, en lugar de Doer
	testUser := &user.User{Doer: mockDoer}

	mockDoer.EXPECT().
		DoSomething(gomock.Any(), gomock.Any()).
		Return(nil).
		Do(func(x int, y string) {
			t.Log("Llamado con x =", x, " e y =", y)
			if x > len(y) {
				//Hacemos que el caso falle
				t.Fail()
			}
		})

	testUser.Use()
}

func TestAccionesBis(t *testing.T) {
	//Controlador del test
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	//Crea una instancia del mock
	mockDoer := mocks.NewMockDoer(mockCtrl)
	//Crea una instancia del tipo a probar, usando el mock de Doer, en lugar de Doer
	testUser := &user.User{Doer: mockDoer}

	mockDoer.EXPECT().
		DoSomething(gomock.Any(), gomock.Any()).
		Return(nil).
		Do(func(x int, y string) {
			t.Log("Llamado con x =", x, " e y =", y)
			if x != 123 {
				//Hacemos que el caso falle
				t.Fail()
			}
		})

	testUser.Use()
}
