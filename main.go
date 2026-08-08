package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/snc-software/go-template-service/docs"
	"github.com/snc-software/go-template-service/domain/services"
	"github.com/snc-software/go-template-service/routes"
	"github.com/snc-software/go-template-service/routes/middleware"
)

// @title           Template API
// @version         1.0
// @description     A basic template management API
// @host            localhost:8080
// @BasePath        /
// @tag.name        Templates
// @tag.description Operations for managing templates
func main() {
	godotenv.Load(".env")
	godotenv.Overload(".env.local")
	templateService := services.TemplateService{}

	router := chi.NewRouter()
	router.Use(middleware.ErrorHandler)
	router.Mount("/templates", routes.TemplateRoutes(templateService))
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	fmt.Println("Server running on http://localhost:8080/swagger/index.html")
	http.ListenAndServe(":8080", router)
}
