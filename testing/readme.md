mockgen tiene dos modos de operación: source y reflect.

El modo source genera mock interfaces a partir del fuente. Se utiliza este modo cuando se usa el flag `-source`:

```ps
mockgen -source .\doer\doer.go -destination .\mocks\mock_doer.go -package mocks
```

- Con `destination` indicamos donde se debe guardar el mock generado. El directorio se tiene que crear antes
- Con `package` decimos que el mock generado se cree en el package `mocks`
- Con `source` indicamos que el paquete que debemos mockear es `./doer/doer.go`

Creara un interface para mockear nuestro interface `Doer`: 

```go
type MockDoer struct {
```

En el código generado se crea un método `EXPECT()` que usaremos en los tests. Se crea un interface `MockDoer` que sirve para mockear nuestro interface `Doer`. Se crea también un método que nos permitira crear una instancia del mock, `func NewMockDoer(ctrl *gomock.Controller) *MockDoer {`.

Con esto, para hacer el test usando el mock de Doer sería:

```go
func TestUse(t *testing.T) {
```

Creamos el controlador:

```go
//Controlador del test
mockCtrl := gomock.NewController(t)
defer mockCtrl.Finish()
```

Creamos una instancia del mock:

```go
//Crea una instancia del mock
mockDoer := mocks.NewMockDoer(mockCtrl)
```

Supongamos que esperamos que el método `DoSomething` se le llame con los argumentos `123` y ` and `"Hello GoMock"` y que devuelva `nil`. Podemos especificar el número de veces que se va a llamar a este método:

- Numero de veces que se llamara: `.Times(number)`
- Máximo número de veces: `.MaxTimes(number)`
- Número mínimo de veces: `.MinTimes(number)`

Ejecutamos la lógica que queremos probar, y especificamos las assertions que queremos verificar:

```go
//Crea una instancia del tipo a probar, usando el mock de Doer, en lugar de Doer
testUser := &user.User{Doer: mockDoer}

//Especifica las assertions
mockDoer.EXPECT().DoSomething(123, "Hello GoMock").Return(nil).Times(1)

testUser.Use()
```

Las assertions se validarán cuando se ejecute `defer mockCtrl.Finish()`.
