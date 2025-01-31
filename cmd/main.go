package main

import (
	"context"
	todo_app "github.com/che1nov/todo-app"
	"github.com/che1nov/todo-app/pkg/handler"
	"github.com/che1nov/todo-app/pkg/repository"
	"github.com/che1nov/todo-app/pkg/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := initConfig(); err != nil {
		logrus.Fatalf("ошибка инициализации параметров: %s", err.Error())
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
	handlers := handler.NewHandler(services)

	srv := new(todo_app.Server)
	go func() {
		if err = srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("ошибка при запуске http-сервера: %s", err.Error())
		}
	}()
	logrus.Print("TodoApp запущено")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("Всё!")
	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("ошибка завершения работы сервера", err.Error())
	}
	if err := db.Close(); err != nil {
		logrus.Errorf("Упс!")
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
