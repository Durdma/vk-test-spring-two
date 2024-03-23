package flood_control

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"task/config"
	"time"
)

type FloodControl struct {
	maxNumberOfRequests int64         // Максимальное количество запросов в единицу времени
	maxNumberOfRetries  int           // Максимальное количество попыток обращения к БД
	floodInterval       time.Duration // Время, за которое не должно быть превышений лимита вызовов
	retryInterval       time.Duration // Время, через которое должен происходить повторный запрос к БД

	redisPool *redis.Pool // Пул соединений с БД
}

// NewFloodControl - Инициализация сущности FloodControl
func NewFloodControl(cfg config.FloodControlConfig, redisPool *redis.Pool) *FloodControl {
	return &FloodControl{
		maxNumberOfRequests: cfg.MaxNumberOfRequests,
		maxNumberOfRetries:  cfg.MaxNumberOfRetries,
		floodInterval:       cfg.FloodControlTTL,
		retryInterval:       cfg.TimeInterval,
		redisPool:           redisPool,
	}
}

// Check - флуд-контроль
func (fc *FloodControl) Check(ctx context.Context, userId int64) (bool, error) {
	key := fmt.Sprintf("user:%v", userId) //Создание ключа для БД
	rs := redsync.New(redigo.NewPool(fc.redisPool))

	mutex := rs.NewMutex(fmt.Sprintf("mutex:%v", key)) //Создание именованного мьютекса
	for i := 0; i < fc.maxNumberOfRetries; i++ {       //Цикл подключения к БД, если мьютекс заблокирован
		if err := mutex.Lock(); err != nil {
			if err != nil && strings.Contains(err.Error(), "lock already taken") {
				time.Sleep(fc.retryInterval)
				continue
			}

			return false, err
		}
		break
	}
	defer mutex.Unlock()

	conn := fc.redisPool.Get()
	defer conn.Close()

	reqs, err := redis.String(conn.Do("GET", key)) // Получение значения по ключу из БД
	if err != nil && !errors.Is(err, redis.ErrNil) {
		return false, err
	}

	// Если запись с ключом отсутствует, то создается новая запись
	if errors.Is(err, redis.ErrNil) {
		_, err := conn.Do("INCR", key) //Создание новой записи ключ-значение
		if err != nil {
			return false, err
		}

		_, err = conn.Do("EXPIRE", key, int64(fc.floodInterval.Seconds())) //Установка времени, за которое не должно быть превышений лимит вызовов
		if err != nil {
			return false, err
		}

		return true, nil
	}

	var count int64

	// Если запись существовала, то получаем значение по ключу и преобразуем его в int
	count, err = strconv.ParseInt(reqs, 10, 64)
	if err != nil {
		return false, err
	}

	// Если полученное значение + 1 превышает лимит вызовов, возвращаем false и ошибку
	if count+1 > fc.maxNumberOfRequests {
		return false, errors.New(fmt.Sprintf("limit of requests is exceeded for user: %v", userId))
	}

	if _, err := conn.Do("INCR", key); err != nil {
		return false, err
	}

	return true, nil
}
