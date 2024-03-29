package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jchou8/sling/handlers"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"

	// Postgres database driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// RunServerCmd is the command used to run the server.
var RunServerCmd = &cobra.Command{
	Use:   "runserver",
	Short: "run user authentication server",
	RunE:  runServer,
}

func runServer(cmd *cobra.Command, args []string) error {
	conn, err := gorm.Open("postgres", pgAddr)
	if err != nil {
		log.Fatalf("failed to open DB conn: %s", err.Error())
	}

	srv := echo.New()

	srv.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "HTTP[${time_rfc3339}] ${method} ${path} status=${status} latency=${latency_human}\n",
		Output: io.MultiWriter(os.Stdout),
	}))

	srv.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	srv.File("/", "frontend/build")
	srv.Static("/static", "frontend/build/static")
	srv.POST("/api/register", handlers.NewUserHandler(conn))
	srv.POST("/api/login", handlers.LoginHandler(conn))

	users := srv.Group("api/users")
	users.Use(handlers.NewTokenAuthMiddleware(conn))
	users.GET("/", handlers.GetUsersHandler(conn))
	users.GET("/current", handlers.GetCurrentUserHandler(conn))

	//srv.GET("/api/rooms", handlers.GetRoomsHandler(conn), handlers.NewTokenAuthMiddleware(conn))

	fmt.Println("Listening at localhost:8888...")
	if err := srv.Start(":8888"); err != nil {
		return err
	}

	return nil
}
