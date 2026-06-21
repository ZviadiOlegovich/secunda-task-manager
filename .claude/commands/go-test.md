# Go Unit Test Writer

Write or fix Go unit tests for the specified function, file, or package.

## Usage

```
/go-test <target>
```

`<target>` — функция, файл или пакет для которого нужны тесты.

---

## Rules (follow strictly)

### Naming & structure

- Test file: `<filename>_test.go`, package `package <same>`
- Function: `TestFunctionName(t *testing.T)`
- Subtest: `t.Run("descriptive name", func(t *testing.T) { ... })`
- Helper: всегда начинать с `t.Helper()`

### Table-driven tests — canonical approach

Always use a struct slice + loop for functions with multiple cases:

```go
func TestFoo(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr error
    }{
        {name: "success", input: ..., want: ...},
        {name: "invalid input", input: ..., wantErr: ErrInvalid},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Foo(tt.input)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("want err %v, got %v", tt.wantErr, err)
            }
            if tt.wantErr == nil && got != tt.want {
                t.Errorf("want %v, got %v", tt.want, got)
            }
        })
    }
}
```

### Error checking

- Всегда `errors.Is(err, tt.wantErr)` — не сравнивать строки ошибок
- `t.Errorf` — продолжает тест, собирает все ошибки
- `t.Fatalf` — только когда продолжение теста бессмысленно (nil pointer и т.п.)
- Никогда не писать `if err != nil && tt.wantErr == nil` вместо `errors.Is`

### Mocks

- Только через интерфейсы — struct с полями-функциями:

```go
type mockRepo struct {
    getFn func(ctx context.Context, id int64) (*Entity, error)
}
func (m *mockRepo) Get(ctx context.Context, id int64) (*Entity, error) {
    return m.getFn(ctx, id)
}
```

- Не использовать сторонние mock-библиотеки (mockery, gomock) — только интерфейсы
- Для Redis использовать `miniredis`, для БД — интерфейс

### Helpers & cleanup

- `t.Helper()` в каждой вспомогательной функции
- `t.Cleanup(func() { ... })` вместо `defer` для очистки ресурсов
- `miniredis.RunT(t)` — автоматически останавливается по `t.Cleanup`

### Parallel

- `t.Parallel()` внутри subtests когда тесты независимы и не используют общий стейт

### What NOT to do

- Не тестировать приватные детали реализации — только публичный контракт
- Не использовать `assert`/`require` из testify — только стандартный `testing`
- Не создавать тесты которые всегда зелёные (проверь что тест падает при сломанной логике)
- Не дублировать логику продакшн-кода в тестах
- Не оставлять неиспользуемые возвращаемые значения из helper-функций

### Project-specific conventions

- Моки определяются в том же `_test.go` файле где используются
- Переиспользуемые моки — пакетные переменные в начале файла: `var okRepo = &mockRepo{...}`
- Sentinel ошибки (`ErrNotFound`, `ErrNotMember`) — через `errors.Is`, никогда string match
- Интеграционные тесты — только в `tests/integration/`, не смешивать с unit
- Строковые указатели в тестах: `func strPtr(s string) *string { return &s }`

---

## Steps to execute

1. Прочитай целевой файл и его интерфейсы/зависимости
2. Определи все публичные методы без покрытия или с неполным покрытием
3. Для каждого метода составь таблицу кейсов: happy path + все ошибочные пути
4. Напиши тесты строго по правилам выше
5. Запусти `go test ./...` и убедись что все проходят
6. Убедись что при намеренно сломанной логике тест падает
