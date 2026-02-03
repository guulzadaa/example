package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"bookstore/internal/logic"
	"bookstore/internal/models"
)

type FrontendHandler struct {
	templates *template.Template
	bookSvc   *logic.BookService
	authSvc   *logic.AuthService
}

type PageData struct {
	Title string
	Books []models.Book
}

func NewFrontendHandler(bookSvc *logic.BookService, authSvc *logic.AuthService) (*FrontendHandler, error) {
	tmpl, err := template.ParseGlob(filepath.Join("web", "templates", "*.html"))
	if err != nil {
		return nil, err
	}

	return &FrontendHandler{
		templates: tmpl,
		bookSvc:   bookSvc,
		authSvc:   authSvc,
	}, nil
}

func (h *FrontendHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Catalog",
		Books: h.bookSvc.ListBooks(),
	}
	h.render(w, "catalog.html", data)
}

func (h *FrontendHandler) Login(w http.ResponseWriter, r *http.Request) {
	data := PageData{Title: "Login"}
	h.render(w, "login.html", data)
}

func (h *FrontendHandler) Register(w http.ResponseWriter, r *http.Request) {
	data := PageData{Title: "Register"}
	h.render(w, "register.html", data)
}

func (h *FrontendHandler) AdminBooks(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Admin | Books",
		Books: h.bookSvc.ListBooks(),
	}
	h.render(w, "admin_books.html", data)
}

func (h *FrontendHandler) RegisterPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	if err := h.authSvc.Register(email, password); err != nil {
		http.Error(w, "register error: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *FrontendHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	token, err := h.authSvc.Login(email, password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *FrontendHandler) render(w http.ResponseWriter, name string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
