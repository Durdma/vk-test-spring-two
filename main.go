package main

import (
	"context"
	"log"
	"task/config"
	"task/flood_control"
	"task/redisdb"
)

const configPath = "config/config.yml"

func main() {
	// Здесь показан пример инициализации флуд-контроля
	// Чтение файла конфигурации
	cfg, err := config.Init(configPath)
	if err != nil {
		log.Fatal(err)
	}

	pool := redisdb.NewRedisPool(cfg.Redis)                               // Создание пула подключений к БД
	floodControl := flood_control.NewFloodControl(cfg.FloodControl, pool) // Создание сущности флуд-контроль

	//Пример вывзова функции Check
	resp, err := floodControl.Check(context.Background(), 111)
	if err != nil {
		log.Println("resp: ", resp, "err: ", err)
		return
	} else {
		log.Println("resp: ", resp)
	}
}

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}
