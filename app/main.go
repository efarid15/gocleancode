package app

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"gocleancode/article/delivery/http"
	"gocleancode/article/delivery/http/middleware"
	mysql2 "gocleancode/article/repository/mysql"
	"gocleancode/article/usecase"
	"gocleancode/author/repository/mysql"
	"log"
	"net/url"
	"time"
)

func init() {
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Println("Service RUN on DEBUG mode")
	}
}

func main() {
	dbHost := viper.GetString(`database.host`)
	dbPort := viper.GetString(`database.port`)
	dbUser := viper.GetString(`database.user`)
	dbPass := viper.GetString(`database.pass`)
	dbName := viper.GetString(`database.name`)
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	dbConn, err := sql.Open(`mysql`, dsn)

	if err != nil {
		log.Fatal(err)
	}
	err = dbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	e := echo.New()
	middL := middleware.InitMiddleware()
	e.Use(middL.CORS)
	authorRepo := mysql.NewMysqlAuthorRepository(dbConn)
	ar := mysql2.NewMysqlArticleRepository(dbConn)

	timeoutContext := time.Duration(viper.GetInt("context.timeout")) * time.Second
	au := usecase.NewArticleUsecase(ar, authorRepo, timeoutContext)
	http.NewArticleHandler(e, au)

	log.Fatal(e.Start(viper.GetString("server.address")))
}