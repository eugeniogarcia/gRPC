# Mocks

Vamos a utilizar mockgen para generar mocks. Mockgen tiene dos modos de operación: source y reflect.

## Source

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

## Matchers

Cuando queramos usar expresiones para definir los argumentos de un mock, gomock nos da la posibilidad. Además de poder especificar un valor concreto como hemos hecho en el ejemplo anterior, podemos usar un Matcher:

- `gomock.Any()`: representa cualquier valor de cualquier tipo
- `gomock.Eq(x)`: usa reflection para matchear valores que sean iguales que x
- `gomock.Nil()`: matchea `nil`
- `gomock.Not(m)`: donde m es otro matcher, matchea cualquier valor que no corresponda con m
- `gomock.Not(x)`: donde x _no es otro matcher_, matchea valores que no sean iguales, campo a campo, con x

Si queremos crear un mock independientemente de lo que valga el primer argumento, pondríamos:

```go
mockDoer.EXPECT().DoSomething(gomock.Any(), "Hello GoMock")
```

Se pueden crear matchers custom implementando este interface:

```go
type Matcher interface {
    Matches(x interface{}) bool
    String() string
}
```

En el paquete match hemos definido un matcher custom llamado `deTipo` . Lo usamos en nuestro caso de prueba:

```go
mockDoer.EXPECT().DoSomething(123, miMatcher.EsTipo("string")).Return(nil).Times(1)
```

## Orden de las llamadas

Si necesitamos que las llamadas al mock se hagan en un determinado orden, podemos usar `.After`:

```go
llamaPrimero := mockDoer.EXPECT().DoSomething(0, "first this")
llamaA := mockDoer.EXPECT().DoSomething(1, "then this").After(llamaPrimero)
_ = mockDoer.EXPECT().DoSomething(2, "or this").After(llamaPrimero)
_ = mockDoer.EXPECT().DoSomething(3, "finally").After(llamaA)
```

Sería equivalente a esta construcción:

```go
gomock.InOrder(
    mockDoer.EXPECT().DoSomething(0, "first this"),
    mockDoer.EXPECT().DoSomething(1, "then this"),
    mockDoer.EXPECT().DoSomething(2, "or this"),
    mockDoer.EXPECT().DoSomething(3, "finally"),
)
```

## Acciones

Podemos especificar un lambda asociado a la ejecución de un mock. Por ejemplo:

```go
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
```

## Generación Automática

Para automatizar la generación de los mocks podemos usar __go generate__. Basta con añadir un comentario en los interfaces para los que necesitemos generar el mock. Por ejemplo, en nuestro caso:

```go
package doer

//go:generate mockgen -source .\doer.go -destination ..\mocks\mock_doer.go -package mocks

//Doer interface a mockear
type Doer interface {
	DoSomething(int, string) error
}
```

Si en el raíz del proyecto ejecutamos:

```ps
go generate ./...
```

Se generarán todos los mocks que hayamos anotado con el comentario `//go:generate mockgen ...`.