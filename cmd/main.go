package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Request структура для входящего запроса
type Request struct {
	Number1   float64
	Number2   float64
	Operation string
}

// Operation структура для хранения истории операций
type Operation struct {
	ID        int
	Number1   float64
	Number2   float64
	Operation string
	Result    float64
	Message   string
	Timestamp time.Time
}

// success - константа успешного выполнения операции
const success = "Ответ рассчитан"

// errSaveOperation - константа сообщения об ошибке сохранения
const errSaveOperation = "Ошибка сохранения: %v\n"

// calculateUsecase выполняет вычисление на основе запроса
func calculateUsecase(req Request) (string, float64) {
	var msg string
	var result float64

	// Выполняем операцию
	switch req.Operation {
	case "+":
		msg = success
		result = req.Number1 + req.Number2
	case "-":
		msg = success
		result = req.Number1 - req.Number2
	case "*":
		msg = success
		result = req.Number1 * req.Number2
	case "/":
		if req.Number2 == 0 {
			msg = "Деление на ноль невозможно"
			result = 0
		} else {
			msg = success
			result = req.Number1 / req.Number2
		}
	default:
		msg = "Неизвестная операция: " + req.Operation
		result = 0
	}

	return msg, result
}

// saveOperation сохраняет операцию в базу данных
func saveOperation(db *sql.DB, req Request, msg string, result float64) error {
	query := `
		INSERT INTO operations (number1, number2, operation, result, message, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var id int
	err := db.QueryRow(query, req.Number1, req.Number2, req.Operation, result, msg, time.Now()).Scan(&id)
	if err != nil {
		return fmt.Errorf("ошибка сохранения операции: %w", err)
	}

	log.Printf("Операция сохранена в БД с ID: %d\n", id)
	return nil
}

// getHistory получает историю операций из базы данных
func getHistory(db *sql.DB) ([]Operation, error) {
	query := `
		SELECT id, number1, number2, operation, result, message, timestamp
		FROM operations
		ORDER BY timestamp DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории: %w", err)
	}
	defer rows.Close()

	var operations []Operation
	for rows.Next() {
		var op Operation
		err := rows.Scan(&op.ID, &op.Number1, &op.Number2, &op.Operation, &op.Result, &op.Message, &op.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		operations = append(operations, op)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации строк: %w", err)
	}

	return operations, nil
}

// connectDB подключается к базе данных PostgreSQL
func connectDB() (*sql.DB, error) {
	connStr := "host=localhost port=5433 user=postgres password=postgres dbname=lizzycalc sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка ping БД: %w", err)
	}

	log.Println("Подключение к базе данных установлено")
	return db, nil
}

// createTable создаёт таблицу operations если её нет
func createTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS operations (
			id SERIAL PRIMARY KEY,
			number1 DOUBLE PRECISION NOT NULL,
			number2 DOUBLE PRECISION NOT NULL,
			operation VARCHAR(10) NOT NULL,
			result DOUBLE PRECISION NOT NULL,
			message VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP NOT NULL
		)`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы: %w", err)
	}

	log.Println("Таблица operations создана или уже существует")
	return nil
}

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаём таблицу
	if err := createTable(db); err != nil {
		log.Fatal(err)
	}

	// Создаём запросы вручную
	req1 := Request{Number1: 10, Number2: 5, Operation: "+"}
	msg, result := calculateUsecase(req1)
	if msg == success {
		log.Printf("%.2f %s %.2f = %.2f (%s)\n", req1.Number1, req1.Operation, req1.Number2, result, msg)
	} else {
		log.Printf("Ошибка: %s\n", msg)
	}
	if err := saveOperation(db, req1, msg, result); err != nil {
		log.Printf(errSaveOperation, err)
	}
	/*
		req2 := Request{Number1: 20, Number2: 8, Operation: "-"}
		msg, result = calculateUsecase(req2)
		if msg == success {
			log.Printf("%.2f %s %.2f = %.2f (%s)\n", req2.Number1, req2.Operation, req2.Number2, result, msg)
		} else {
			log.Printf("Ошибка: %s\n", msg)
		}
		if err := saveOperation(db, req2, msg, result); err != nil {
			log.Printf(errSaveOperation, err)
		}

		req3 := Request{Number1: 7, Number2: 3, Operation: "*"}
		msg, result = calculateUsecase(req3)
		if msg == success {
			log.Printf("%.2f %s %.2f = %.2f (%s)\n", req3.Number1, req3.Operation, req3.Number2, result, msg)
		} else {
			log.Printf("Ошибка: %s\n", msg)
		}
		if err := saveOperation(db, req3, msg, result); err != nil {
			log.Printf(errSaveOperation, err)
		}

		req4 := Request{Number1: 15, Number2: 3, Operation: "/"}
		msg, result = calculateUsecase(req4)
		if msg == success {
			log.Printf("%.2f %s %.2f = %.2f (%s)\n", req4.Number1, req4.Operation, req4.Number2, result, msg)
		} else {
			log.Printf("Ошибка: %s\n", msg)
		}
		if err := saveOperation(db, req4, msg, result); err != nil {
			log.Printf(errSaveOperation, err)
		}

		req5 := Request{Number1: 10, Number2: 0, Operation: "/"}
		msg, result = calculateUsecase(req5)
		if msg == success {
			log.Printf("%.2f %s %.2f = %.2f (%s)\n", req5.Number1, req5.Operation, req5.Number2, result, msg)
		} else {
			log.Printf("Ошибка: %s\n", msg)
		}
		if err := saveOperation(db, req5, msg, result); err != nil {
			log.Printf(errSaveOperation, err)
		}
	*/
	// Получаем и выводим историю из базы данных
	log.Println("\n=== История операций из БД ===")
	history, err := getHistory(db)
	if err != nil {
		log.Fatal(err)
	}

	if len(history) == 0 {
		log.Println("История операций пуста")
	} else {
		for _, op := range history {
			log.Printf("\nОперация #%d:\n", op.ID)
			log.Printf("  Число 1: %.2f\n", op.Number1)
			log.Printf("  Число 2: %.2f\n", op.Number2)
			log.Printf("  Операция: %s\n", op.Operation)
			log.Printf("  Результат: %.2f\n", op.Result)
			log.Printf("  Статус: %s\n", op.Message)
			log.Printf("  Время: %s\n", op.Timestamp.Format("2006-01-02 15:04:05"))
		}
	}
	log.Println("\n==============================")
}
