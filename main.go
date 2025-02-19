package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go-gorm/models"
	"go-gorm/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Title     string `json:"title"`
	Author    string `json:"author"`
	Publisher string `json:"publisher"`
}

type Repo struct {
	DB *gorm.DB
}

func (r *Repo) CreateBook(ctx *fiber.Ctx) error {
	book := Book{}
	err := ctx.BodyParser(&book)

	if err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}
	err = r.DB.Create(&book).Error
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Unable to create book"})
		return err
	}

	ctx.Status(http.StatusOK).JSON(
		&fiber.Map{"message": "Book Created"})
	return nil
}

func (r *Repo) GetAllBooks(ctx *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "Unable to Get All books"})
		return err
	}

	ctx.Status(http.StatusOK).JSON(
		&fiber.Map{
			"message": "Book Fetched successfully",
			"data":    bookModels,
		})
	return nil
}

func (r *Repo) DeleteBook(ctx *fiber.Ctx) error {
	bookModel := &models.Books{}
	id := ctx.Params("id")

	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	err := r.DB.Delete(bookModel, id)

	if err.Error != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete book",
		})
		return err.Error
	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book delete successfully",
	})
	return nil
}

func (r *Repo) GetBookByID(ctx *fiber.Ctx) error {

	id := ctx.Params("id")
	bookModel := &models.Books{}
	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}

	fmt.Println("the ID is", id)

	err := r.DB.Where("id = ?", id).First(bookModel).Error

	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get the book"})
		return err
	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book id fetched successfully",
		"data":    bookModel,
	})

	return nil
}

func (r *Repo) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_book", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_book/:id", r.GetBookByID)
	api.Get("/books", r.GetAllBooks)

}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("Cannot load database")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	repo := Repo{
		DB: db,
	}
	app := fiber.New()
	repo.SetupRoutes(app)

	app.Listen(":8080")

}
