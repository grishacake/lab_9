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
	dbname = "query"
)

type Handlers struct {
	dbProvider *DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

// Метод для получения приветствия по имени
func (h *Handlers) GetGreeting(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Нет параметра 'name'"})
	}

	greeting, err := h.dbProvider.SelectGreeting(name)
	if err != nil {
		c.Logger().Error("Ошибка при получении приветствия: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении приветствия"})
	}

	return c.String(http.StatusOK, greeting)
}

// Методы для работы с базой данных

// SelectGreeting получает приветствие из базы данных
func (dp *DatabaseProvider) SelectGreeting(name string) (string, error) {
	var greeting string
	row := dp.db.QueryRow("SELECT greeting FROM greetings WHERE name = $1", name)
	err := row.Scan(&greeting)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если записи нет, создаем новое приветствие для данного имени
			_, err := dp.db.Exec("INSERT INTO greetings (name, greeting) VALUES ($1, $2)", name, fmt.Sprintf("Hello, %s!", name))
			if err != nil {
				return "", err
			}
			greeting = fmt.Sprintf("Hello, %s!", name)
		} else {
			return "", err
		}
	}
	return greeting, nil
}

func main() {
	// Формируем строку подключения к базе данных
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)

	// Открываем соединение с базой данных
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}
	defer db.Close()

	// Проверяем соединение с базой данных
	if err := db.Ping(); err != nil {
		log.Fatal("Ошибка пинга базы данных:", err)
	}

	// Создаем провайдер для работы с базой данных
	dp := &DatabaseProvider{db: db}
	// Создаем обработчики
	h := &Handlers{dbProvider: dp}

	// Создаем новый экземпляр Echo
	e := echo.New()

	// Регистрируем маршруты
	e.GET("/api/user", h.GetGreeting)

	// Запуск сервера на порту 8081
	e.Logger.Fatal(e.Start(":8081"))
}
