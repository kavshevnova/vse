# Honey-task

Коллекция практических задач по различным темам golang которые помогут закрепить знания.

## Доступные темы

### default block
- **[slices](slices/)** - Слайсы в Go (50 задач)
- **[maps](maps/)** - Maps в Go (50 задач)
- **[pointers](pointers/)** - Указатели в Go (50 задач)
- **[structs](structs/)** - Структуры в Go (50 задач)
- **[interface](interface/)** - Интерфейсы в Go (30 задач)
- **[algo](algo/)** - Алго задачи с собеседований Go(10 задач)
- _defer|panic|recover_ - механизмы восстановления (soon)
- _errors_ - ошибки (soon)
- _strings_ - строки (its needed?)
- _func_ - функции (its needed?)

### concurrency block
- **[concurrency](concurrency/)** - Конкурентность в Go (30 задач)

### Practice block
- **[code-review](code-review)** - Код ревью задачи с собеседований(20 задач)
### Algo block
- **[algo](algo/)** - Алго задачи(10 задач)

## Использование

```bash
# Смотри список доступных тем
go run main.go list

# Справка по теме
go run main.go <тема>

# Запуск задачи
go run main.go <тема> <номер>

# Примеры:
go run main.go slices 3           # задача 3 по слайсам
go run main.go slices 1 5 10      # несколько задач
go run main.go concurrency 1      # задача 1 по конкурентности
```

### Для concurrency задач:

```bash
# Вариант 1: Напрямую из папки задачи
cd concurrency/task001/
go run main.go
# 1. Изучить код и найти ошибку
cat main.go

# 2. Попробовать запустить (увидеть ошибку)
go run main.go

# 3. Исправить код
nano main.go

# 4. Проверить исправление
go run main.go
```

### Для interface задач:
```bash
# Вариант 1: Напрямую из папки задачи
cd interface/task001/
cat main.go  # изучить интерфейсы
# Реализовать требуемые интерфейсы и структуры

# Вариант 2: Через main.go
go run main.go interface 1
```

### Для review задач:
```bash
cd code-review/task001
go run main.go
# 1. Изучить код и найти ошибку
cat main.go

# 2. Исправить код
nano main.go

# 4. Проверить исправление
go run main.go
```

## Почитать полезное:

- [Go Documentation](https://go.dev/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Blog](https://go.dev/blog/)
