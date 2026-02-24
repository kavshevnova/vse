# Задача: решение всех задач для подготовки к собеседованию по Go

## Дата: 22.02.2025

## Контекст
Репозиторий с практическими задачами по Go для подготовки к собеседованиям.

## Категории задач

### 1. maps/tasks.go — 50 задач "Что выведет?"
Анализ поведения map в Go: nil map, zero values, итерация, передача по ссылке, concurrent writes и др.

### 2. slices/tasks.go — 50 задач "Что выведет?"
Анализ поведения slice: len/cap, append, copy, sub-slices, три-индексная нарезка и др.

### 3. structs/tasks.go — 50 задач "Что выведет?"
Структуры: embedding, json теги, pointer receivers, value receivers, unexported fields и др.

### 4. pointers/tasks.go — 50 задач "Что выведет?"
Указатели: nil pointer, pointer arithmetic, escape analysis, double pointers и др.

### 5. algo/task001-010 — Алгоритмические задачи
Реализация алгоритмов: группировка, анаграммы, дерево, граф и др.

### 6. concurrency/task001-020 — Реализация
Горутины, каналы: merge channels, worker pool, pipeline, circuit breaker и др.

### 7. concurrency/task021-030 — "Что выведет?"
Race conditions, select, defer, nil interfaces и др.

### 8. interface/task001-030 — Реализация паттернов
Producer/Consumer, Cache, EventBus, RateLimiter, WorkerPool, Repository и др.

### 9. code-review/task001-021 — Ревью кода
Нахождение багов: SQL injection, race condition, утечки ресурсов и др.

## План реализации

1. Запустить "что выведет" задачи для получения точных ответов
2. Аннотировать все "что выведет" задачи комментариями с ответами
3. Реализовать алгоритмические задачи
4. Реализовать concurrency задачи
5. Реализовать interface задачи
6. Написать code review комментарии
7. Запустить сборку и линтер

## Подход к аннотации
Добавляем комментарий `// OUTPUT:` перед каждой функцией с ожидаемым выводом.
