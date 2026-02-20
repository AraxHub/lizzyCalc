package calculator

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"lizzyCalc/internal/domain"
	"lizzyCalc/internal/mocks"
)

// newTestLogger создаёт логгер для тестов (выводит только ошибки, чтобы не засорять вывод).
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}


// Тест 1: Cache Hit — результат берётся из кэша, БД не вызывается
func TestCalculate_CacheHit(t *testing.T) {
	// Создаём контроллер gomock — он управляет жизненным циклом моков,
	// отслеживает вызовы и проверяет, что все ожидания выполнены.
	ctrl := gomock.NewController(t)
	// defer ctrl.Finish() вызовется в конце теста и проверит:
	// - все ли EXPECT() были вызваны
	// - не было ли лишних вызовов
	defer ctrl.Finish()

	// Создаём моки всех зависимостей UseCase.
	// Это не настоящие реализации — это "актёры", которые будут
	// отвечать заранее запрограммированными ответами.
	mockCache := mocks.NewMockICache(ctrl)
	mockRepo := mocks.NewMockIOperationRepository(ctrl)
	mockBroker := mocks.NewMockIProducer(ctrl)
	mockAnalytics := mocks.NewMockIOperationAnalytics(ctrl)

	// Программируем поведение мока кэша:
	// "Когда вызовут Get с любым ctx и ключом '10 + 5',
	//  притворись, что нашёл значение и верни (15.0, true, nil)"
	// Мок НЕ хранит данные — он просто возвращает то, что мы указали.
	mockCache.EXPECT().
		Get(gomock.Any(), "10 + 5"). // gomock.Any() — любой context
		Return(15.0, true, nil)      // value=15.0, found=true, err=nil

	// Создаём тестируемый UseCase, передавая ему моки вместо реальных зависимостей.
	// UseCase не знает, что это моки — он работает с ними как с обычными интерфейсами.
	uc := New(mockRepo, mockCache, mockBroker, mockAnalytics, newTestLogger())

	// Вызываем тестируемый метод.
	// Внутри Calculate вызовется u.cache.Get(ctx, "10 + 5"),
	// мок перехватит этот вызов и вернёт (15.0, true, nil).
	// Calculate увидит found=true и сразу вернёт результат из "кэша",
	// не обращаясь к repo, broker и т.д.
	result, err := uc.Calculate(context.Background(), 10, 5, "+")

	// Проверяем результат.
	// require.NoError — если err != nil, тест сразу падает (нет смысла проверять result).
	require.NoError(t, err)
	// assert.Equal — проверяет равенство. Если не равны, тест помечается как failed,
	// но продолжает выполняться (увидим все упавшие проверки).
	assert.Equal(t, 15.0, result.Result)   // результат из "кэша"
	assert.Equal(t, 10.0, result.Number1)  // входные данные сохранены
	assert.Equal(t, 5.0, result.Number2)
	assert.Equal(t, "+", result.Operation)
	// После выхода из функции ctrl.Finish() проверит, что Get был вызван ровно 1 раз.
}


// Тест 2: Cache Miss — полный флоу: расчёт → БД → кэш → брокер
func TestCalculate_CacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockICache(ctrl)
	mockRepo := mocks.NewMockIOperationRepository(ctrl)
	mockBroker := mocks.NewMockIProducer(ctrl)
	mockAnalytics := mocks.NewMockIOperationAnalytics(ctrl)

	// gomock.InOrder гарантирует, что вызовы произойдут именно в этом порядке.
	// Если порядок нарушится — тест упадёт.
	gomock.InOrder(
		mockCache.EXPECT().Get(gomock.Any(), "10 + 5").Return(0.0, false, nil),
		mockRepo.EXPECT().SaveOperation(gomock.Any(), gomock.Any()).Return(nil),
		mockCache.EXPECT().Set(gomock.Any(), "10 + 5", 15.0).Return(nil),
		mockBroker.EXPECT().Send(gomock.Any(), []byte("10 + 5"), gomock.Any()).Return(nil),
	)

	uc := New(mockRepo, mockCache, mockBroker, mockAnalytics, newTestLogger())

	result, err := uc.Calculate(context.Background(), 10, 5, "+")

	require.NoError(t, err)
	assert.Equal(t, 15.0, result.Result)
}

// Тест 3: Ошибка — деление на ноль
func TestCalculate_DivisionByZero(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockICache(ctrl)
	mockRepo := mocks.NewMockIOperationRepository(ctrl)
	mockBroker := mocks.NewMockIProducer(ctrl)
	mockAnalytics := mocks.NewMockIOperationAnalytics(ctrl)

	// Кэш-мисс — идём считать
	mockCache.EXPECT().Get(gomock.Any(), "10 / 0").Return(0.0, false, nil)
	// repo, cache.Set, broker НЕ вызываются — ошибка раньше

	uc := New(mockRepo, mockCache, mockBroker, mockAnalytics, newTestLogger())

	result, err := uc.Calculate(context.Background(), 10, 0, "/")

	// Ожидаем ошибку, result должен быть nil
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")
}

// Тест 4: История операций
func TestHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIOperationRepository(ctrl)

	// Готовим данные, которые "вернёт БД"
	expected := []domain.Operation{
		{ID: 1, Number1: 10, Number2: 5, Operation: "+", Result: 15},
		{ID: 2, Number1: 20, Number2: 4, Operation: "/", Result: 5},
	}

	mockRepo.EXPECT().GetHistory(gomock.Any()).Return(expected, nil)

	// Для History не нужны cache, broker, analytics — передаём nil
	uc := New(mockRepo, nil, nil, nil, newTestLogger())

	result, err := uc.History(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expected, result)
}
