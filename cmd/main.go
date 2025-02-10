package main

import (
	"context"
	"github.com/che1nov/todo-app/internal/handlers"
	"github.com/che1nov/todo-app/internal/repository"
	"github.com/che1nov/todo-app/internal/server"
	"github.com/che1nov/todo-app/internal/service"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {
	if err := initConfig(); err != nil {
		logrus.Fatalf("ошибка инициализации конфигурации: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("ошибка загрузки переменных среды: %s", err.Error())
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		logrus.Fatalf("ошибка инициализации базы данных: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handlers.NewHandler(services)

	srv := new(server.Server)
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("ошибка при запуске http-сервера: %s", err.Error())
		}
	}()
	logrus.Print("GeoService запущен")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("завершение работы сервера...")
	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("ошибка завершения работы сервера: %s", err.Error())
	}
	if err := db.Close(); err != nil {
		logrus.Errorf("ошибка при закрытии базы данных: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
