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
	dbname = "lab_8"
)

type Handlers struct {
	dbProvider *DatabaseProvider
}
type DatabaseProvider struct {
	db *sql.DB
}
type CountRequest struct {
	Msg string `json:"msg"`
}

// Методы для работы с базой данных
// SelectHello выбирает случайное сообщение из таблицы hello
func (dp *DatabaseProvider) SelectHello() (string, error) {
	var msg string
	row := dp.db.QueryRow("SELECT message FROM hello ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}
	return msg, nil
}

// InsertHello вставляет новое сообщение в таблицу hello
func (dp *DatabaseProvider) InsertHello(msg string) error {
	_, err := dp.db.Exec("INSERT INTO hello (message) VALUES ($1)", msg)
	if err != nil {
		return err
	}
	return nil
}

// Обработчики HTTP-запросов
// GetHello - обрабатывает GET запрос для получения случайного сообщения
func (h *Handlers) GetHello(c echo.Context) error {
	msg, err := h.dbProvider.SelectHello()
	if err != nil {
		c.Logger().Error("Ошибка при получении сообщения из базы данных:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении данных"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": msg})
}

// PostHello - обрабатывает POST запрос для добавления нового сообщения
func (h *Handlers) PostHello(c echo.Context) error {
	var input CountRequest
	if err := c.Bind(&input); err != nil {
		c.Logger().Error("Ошибка парсинга JSON:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка парсинга JSON"})
	}
	err := h.dbProvider.InsertHello(input.Msg)
	if err != nil {
		c.Logger().Error("Ошибка при вставке сообщения в базу данных:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении сообщения"})
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": input.Msg})
}

// Основная функция
func main() {
	// Формируем строку подключения для PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)
	// Создаем соединение с базой данных
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}
	defer db.Close()
	// Проверка соединения с БД
	if err := db.Ping(); err != nil {
		log.Fatal("Ошибка пинга базы данных:", err)
	}
	// Создаем провайдер для работы с БД
	dp := &DatabaseProvider{db: db}
	// Создаем обработчики
	h := &Handlers{dbProvider: dp}
	// Создаем новый сервер Echo
	e := echo.New()
	// Регистрация маршрутов с обработчиками
	e.GET("/get", h.GetHello)
	e.POST("/post", h.PostHello)
	// Запуск сервера
	e.Logger.Fatal(e.Start(":8081"))
}
