package main

import (
	"fmt"
	"time"
)

// Request структура для входящего запроса
type Request struct {
	Number1   float64
	Number2   float64
	Operation string
}

// Operation структура для хранения истории операций
type Operation struct {
	Number1   float64
	Number2   float64
	Operation string
	Result    float64
	Message   string
	Timestamp time.Time
}

// operations - слайс для хранения истории операций (аналог БД)
var operations []Operation

// success - константа успешного выполнения операции
const success = "Ответ рассчитан"

// calculateUsecase выполняет вычисление на основе запроса и сохраняет в БД
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

	// Сохраняем операцию в БД (слайс)
	operation := Operation{
		Number1:   req.Number1,
		Number2:   req.Number2,
		Operation: req.Operation,
		Result:    result,
		Message:   msg,
		Timestamp: time.Now(),
	}
	operations = append(operations, operation)

	return msg, result
}

// history выводит историю всех операций
func history() {
	if len(operations) == 0 {
		fmt.Println("История операций пуста")
		return
	}

	fmt.Println("\n=== История операций ===")

	for i, op := range operations {
		fmt.Printf("\nОперация #%d:\n", i+1)
		fmt.Printf("  Число 1: %.2f\n", op.Number1)
		fmt.Printf("  Число 2: %.2f\n", op.Number2)
		fmt.Printf("  Операция: %s\n", op.Operation)
		fmt.Printf("  Результат: %.2f\n", op.Result)
		fmt.Printf("  Статус: %s\n", op.Message)
		fmt.Printf("  Время: %s\n", op.Timestamp.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("\n=======================\n")
}

func main() {
	//имитируем приход запросов от клиентов
	req1 := Request{Number1: 10, Number2: 5, Operation: "+"}

	//вызываем бизнес-логику
	msg, result := calculateUsecase(req1)
	if msg == success {
		fmt.Printf("%.2f %s %.2f = %.2f (%s)\n", req1.Number1, req1.Operation, req1.Number2, result, msg)
	} else {
		fmt.Printf("Ошибка: %s\n", msg)
	}

	req2 := Request{Number1: 20, Number2: 8, Operation: "-"}
	msg, result = calculateUsecase(req2)
	if msg == success {
		fmt.Printf("%.2f %s %.2f = %.2f (%s)\n", req2.Number1, req2.Operation, req2.Number2, result, msg)
	} else {
		fmt.Printf("Ошибка: %s\n", msg)
	}

	req3 := Request{Number1: 7, Number2: 3, Operation: "*"}
	msg, result = calculateUsecase(req3)
	if msg == success {
		fmt.Printf("%.2f %s %.2f = %.2f (%s)\n", req3.Number1, req3.Operation, req3.Number2, result, msg)
	} else {
		fmt.Printf("Ошибка: %s\n", msg)
	}

	req4 := Request{Number1: 15, Number2: 3, Operation: "/"}
	msg, result = calculateUsecase(req4)
	if msg == success {
		fmt.Printf("%.2f %s %.2f = %.2f (%s)\n", req4.Number1, req4.Operation, req4.Number2, result, msg)
	} else {
		fmt.Printf("Ошибка: %s\n", msg)
	}

	req5 := Request{Number1: 10, Number2: 0, Operation: "/"}
	msg, result = calculateUsecase(req5)
	if msg == success {
		fmt.Printf("%.2f %s %.2f = %.2f (%s)\n", req5.Number1, req5.Operation, req5.Number2, result, msg)
	} else {
		fmt.Printf("Ошибка: %s\n", msg)
	}

	// Выводим историю всех операций
	history()
}
