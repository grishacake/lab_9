package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "grishka"
	dbname = "count"
)

type DatabaseProvider struct {
	db *sql.DB
}
type CountRequest struct {
	Count int `json:"count"`
}
type Handlers struct {
	dbProvider *DatabaseProvider
}

// Методы для работы с базой данных
// SelectCount выбирает текущее значение счетчика из базы данных
func (dp *DatabaseProvider) SelectCount() (int, error) {
	var count int
	row := dp.db.QueryRow("SELECT count FROM counters WHERE id = 1")
	err := row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если записи нет, создаем начальный счетчик
			_, err := dp.db.Exec("INSERT INTO counters (count) VALUES (0)")
			if err != nil {
				return 0, err
			}
			count = 0
		} else {
			return 0, err
		}
	}
	return count, nil
}

// UpdateCount обновляет значение счетчика в базе данных
func (dp *DatabaseProvider) UpdateCount(increment int) error {
	_, err := dp.db.Exec("UPDATE counters SET count = count + $1 WHERE id = 1", increment)
	if err != nil {
		return err
	}
	return nil
}

// Обработчики HTTP-запросов
// GetCount обрабатывает GET запрос для получения текущего значения счетчика
func (h *Handlers) GetCount(c echo.Context) error {
	count, err := h.dbProvider.SelectCount()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]int{"count": count})
}

// PostCount обрабатывает POST запрос для увеличения счетчика
func (h *Handlers) PostCount(c echo.Context) error {
	var input CountRequest
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка парсинга JSON"})
	}
	if input.Count <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Значение count должно быть положительным числом"})
	}
	err := h.dbProvider.UpdateCount(input.Count)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Счетчик увеличен на %d", input.Count)})
}
func main() {
	// Формирование строки подключения для PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}
	defer db.Close()
	// Проверка соединения
	if err := db.Ping(); err != nil {
		log.Fatal("Ошибка пинга базы данных:", err)
	}
	// Создаем провайдер для работы с БД
	dp := &DatabaseProvider{db: db}
	// Создаем обработчики
	h := &Handlers{dbProvider: dp}
	// Создаем новый сервер Echo
	e := echo.New()
	// Регистрация обработчиков
	e.GET("/count/get", h.GetCount)
	e.POST("/count/post", h.PostCount)
	// Запуск сервера на порту 8081
	e.Logger.Fatal(e.Start(":8081"))
}
