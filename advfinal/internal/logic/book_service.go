package logic

import (
	"errors"

	"bookstore/internal/models"
	"bookstore/internal/repository"
)

type BookService struct {
	repo repository.BookRepository
}

func NewBookService(repo repository.BookRepository) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) ListBooks() []models.Book {
	return s.repo.GetAll()
}

func (s *BookService) GetBook(id int) (models.Book, error) {
	return s.repo.GetByID(id)
}

func (s *BookService) CreateBook(b models.Book) (models.Book, error) {
	if b.Title == "" || b.Author == "" {
		return models.Book{}, errors.New("title and author are required")
	}
	if b.Price < 0 {
		return models.Book{}, errors.New("price cannot be negative")
	}

	return s.repo.Create(b)
}

func (s *BookService) UpdateBook(b models.Book) error {
	if b.ID <= 0 {
		return errors.New("invalid id")
	}
	if b.Title == "" || b.Author == "" {
		return errors.New("title and author are required")
	}
	if b.Price < 0 {
		return errors.New("price cannot be negative")
	}

	return s.repo.Update(b)
}

func (s *BookService) DeleteBook(id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.Delete(id)
}
