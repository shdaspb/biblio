package application

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/text/encoding/charmap"

	"biblio/internal/repository"
)

type app struct {
	ctx   context.Context
	repo  *repository.Repository
	cache map[string]repository.User
}
type BookM struct {
	Author  string
	Name    string
	Message string
}
type UserRole string

var head = filepath.Join("public", "html", "head.html")
var header = filepath.Join("public", "html", "header.html")
var headera = filepath.Join("public", "html", "headera.html")
var pager = filepath.Join("public", "html", "pager.html")

func (a app) Routes(r *httprouter.Router) {
	r.ServeFiles("/public/*filepath", http.Dir("public"))
	r.GET("/", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.LoginPage(rw, "")
	})
	r.POST("/", a.Login)
	r.GET("/logout", a.Logout)
	r.GET("/signup", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupPage(rw, "")
	})
	r.POST("/signup", a.Signup)

	r.GET("/redir", a.Redir)
	r.GET("/user", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "USER" {
			a.StartPage(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))

	r.GET("/user/books", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "USER" {
			a.GetBooks(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))

	r.GET("/user/books/search", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "USER" {
			a.GetBooksSearch(rw, "")
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.POST("/user/books/search", a.PostBooksSearch)

	r.GET("/user/books/open/:id", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "USER" {
			a.GetBooksOpenID(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))

	r.GET("/user/books/read/:id", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "USER" {
			a.GetBooksReadID(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))

	r.GET("/admin", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.StartPagea(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.GET("/admin/users", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.GetUsers(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.GET("/admin/users/delete/:id", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.DeleteUser(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.GET("/admin/users/edit/:id", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.EditUserPage(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.POST("/admin/users/edit/:id", a.EditUser)
	r.GET("/admin/books", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.GetBooksa(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.POST("/admin/books", a.PostBooks)
	r.GET("/admin/books/new", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.AddNewBookPage(rw, "")
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.POST("/admin/books/new", a.AddNewBook)
	r.GET("/admin/books/open/:id", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.GetBooksOpenID(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.GET("/admin/books/read/:id", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.GetBooksReadID(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.GET("/admin/books/delete/:id", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.DeleteBook(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))

	r.GET("/admin/books/edit/:id", a.authorized(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if r.Context().Value("role").(UserRole) == "ADMIN" {
			a.EditBookPage(rw, r, p)
		} else {
			http.Redirect(rw, r, "/redir", http.StatusSeeOther)
		}
	}))
	r.POST("/admin/books/edit/:id", a.EditBook)
}

/*func (a app) authorized(next httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		token, err := readCookie("token", r)

		if err != nil {
			http.Redirect(rw, r, "/", http.StatusSeeOther)
			return
		}

		if _, ok := a.cache[token]; !ok {
			http.Redirect(rw, r, "/", http.StatusSeeOther)
			return
		}
		next(rw, r, ps)
	}
}*/

func (a *app) authorized(next httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		token, err := readCookie("token", r)

		if err != nil {
			http.Redirect(rw, r, "/", http.StatusSeeOther)

			return
		}
		role := UserRole(a.cache[token].Role)
		ctx := context.WithValue(r.Context(), "role", role)
		next(rw, r.WithContext(ctx), ps)
	}
}

func (a app) Redir(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("public", "html", "redir.html")

	tmpl, err := template.ParseFiles(lp, head)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "redir", nil)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) LoginPage(rw http.ResponseWriter, message string) {
	lp := filepath.Join("public", "html", "login.html")

	tmpl, err := template.ParseFiles(lp, head)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "login", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) Login(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	login := r.FormValue("login")
	password := r.FormValue("password")

	if login == "" || password == "" {
		a.LoginPage(rw, "Необходимо указать логин и пароль!")
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	user, err := a.repo.Login(a.ctx, login, hashedPass)
	if err != nil {
		a.LoginPage(rw, "Вы ввели неверный логин или пароль!")
		return
	}
	if !user.Active {
		a.LoginPage(rw, "Пользователь заблокирован!")
		return
	}

	//логин и пароль совпадают, поэтому генерируем токен, пишем его в кеш и в куки
	time64 := time.Now().Unix()
	timeInt := string(time64)
	token := login + password + timeInt

	hashToken := md5.Sum([]byte(token))
	hashedToken := hex.EncodeToString(hashToken[:])

	a.cache[hashedToken] = user

	livingTime := 60 * time.Minute
	expiration := time.Now().Add(livingTime)
	//кука будет жить 1 час
	cookie := http.Cookie{Name: "token", Value: url.QueryEscape(hashedToken), Expires: expiration}
	http.SetCookie(rw, &cookie)
	if user.Role == "ADMIN" {
		http.Redirect(rw, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(rw, r, "/user", http.StatusSeeOther)

	}

}

func (a app) Logout(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1}
		http.SetCookie(rw, &c)
	}
	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

func (a app) SignupPage(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "html", "signup.html")

	tmpl, err := template.ParseFiles(sp, head)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "signup", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) Signup(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	username := strings.TrimSpace(r.FormValue("username"))
	password := strings.TrimSpace(r.FormValue("password"))
	password2 := strings.TrimSpace(r.FormValue("password2"))
	fullName := strings.TrimSpace(r.FormValue("fullName"))

	if username == "" || fullName == "" || password == "" || password2 == "" {
		a.SignupPage(rw, "Все поля должны быть заполнены!")
		return
	}

	if password != password2 {
		a.SignupPage(rw, "Пароли не совпадают! Попробуйте еще")
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	err := a.repo.AddNewUser(a.ctx, username, fullName, hashedPass)
	if err != nil {
		a.SignupPage(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	a.LoginPage(rw, fmt.Sprintf("%s, вы успешно зарегистрированы! Теперь вам доступен вход через страницу авторизации", fullName))
}

func (a app) StartPage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("public", "html", "index.html")

	tmpl, err := template.ParseFiles(lp, head, header)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "index", nil)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}

func (a app) StartPagea(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("public", "html", "index.html")

	tmpl, err := template.ParseFiles(lp, head, headera)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "index", nil)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}
func (a app) GetUsers(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("public", "html", "userslist.html")
	users, err := a.repo.AllUser(a.ctx)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	tmpl, err := template.ParseFiles(lp, head, headera)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "userlist", users)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}

func (a app) DeleteUser(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.DeleteUserById(a.ctx, p.ByName("id"))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(rw, r, "/admin/users", http.StatusSeeOther)

}

func (a app) EditUserPage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sp := filepath.Join("public", "html", "usersedit.html")

	user, err := a.repo.GetUserById(a.ctx, p.ByName("id"))

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	tmpl, err := template.ParseFiles(sp, head, headera)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "usersedit", user)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) EditUser(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var act string
	if r.FormValue("active") == "" {
		act = "false"
	} else {
		act = "true"
	}

	err := a.repo.PutUserById(a.ctx, r.FormValue("role"), act, p.ByName("id"))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(rw, r, "/admin/users", http.StatusSeeOther)
}

func (a app) GetBooksa(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var pages repository.Page
	var err error = nil
	pageNumber := 1
	queryValues := r.URL.Query()

	pageNumberStr := queryValues.Get("page")

	if pageNumberStr != "" {
		// Convert the query parameter to integer
		convertedPageNumber, err := strconv.Atoi(pageNumberStr)
		if err == nil {
			pageNumber = convertedPageNumber
		}
	}
	pages, err = a.repo.AllBook(a.ctx, pageNumber, 12, "", "", "", "", "")
	pages.PageUrl = "/admin/books?page="
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	link := queryValues.Get("link")

	if link != "" {
		pages, err = a.repo.AllBook(a.ctx, pageNumber, 12, "", "", "", "", link)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		pages.PageUrl = "/admin/books?link=" + link + "&page="
	}

	lp := filepath.Join("public", "html", "all-booka.html")

	tmpl, err := template.ParseFiles(lp, head, headera, pager)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "all-booka", pages)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}

func (a app) GetBooks(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var pages repository.Page
	var err error = nil
	pageNumber := 1
	queryValues := r.URL.Query()

	pageNumberStr := queryValues.Get("page")

	if pageNumberStr != "" {
		// Convert the query parameter to integer
		convertedPageNumber, err := strconv.Atoi(pageNumberStr)
		if err == nil {
			pageNumber = convertedPageNumber
		}
	}

	category := queryValues.Get("category")
	author := queryValues.Get("author")
	series := queryValues.Get("series")
	name := queryValues.Get("name")

	if category != "" {
		pages, err = a.repo.AllBook(a.ctx, pageNumber, 12, category, "", "", "", "")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		pages.PageUrl = "/user/books?category=" + category + "&page="
	}

	if author != "" {
		pages, err = a.repo.AllBook(a.ctx, pageNumber, 12, "", author, "", "", "")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		pages.PageUrl = "/user/books?author=" + author + "&page="
	}

	if series != "" {
		pages, err = a.repo.AllBook(a.ctx, pageNumber, 12, "", "", series, "", "")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		pages.PageUrl = "/user/books?series=" + series + "&page="
	}

	if name != "" {
		pages, err = a.repo.AllBook(a.ctx, pageNumber, 12, "", "", "", name, "")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		pages.PageUrl = "/user/books?name=" + name + "&page="
	}

	lp := filepath.Join("public", "html", "all-book.html")

	tmpl, err := template.ParseFiles(lp, head, header, pager)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "all-book", pages)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}

func (a app) PostBooks(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	link := strings.TrimSpace(r.FormValue("link"))

	http.Redirect(rw, r, "/admin/books?link="+link+"&page=1", http.StatusSeeOther)

}

func (a app) GetBooksSearch(rw http.ResponseWriter, message string) {

	sp := filepath.Join("public", "html", "user-search.html")

	tmpl, err := template.ParseFiles(sp, head)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "user-search", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}
func (a app) PostBooksSearch(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	category := strings.TrimSpace(r.FormValue("category"))
	author := strings.TrimSpace(r.FormValue("author"))
	series := strings.TrimSpace(r.FormValue("series"))
	name := strings.TrimSpace(r.FormValue("name"))
	if category != "" && author == "" && series == "" && name == "" {
		http.Redirect(rw, r, "/user/books?category="+category+"&page=1", http.StatusSeeOther)
	} else {
		if category == "" && author != "" && series == "" && name == "" {
			http.Redirect(rw, r, "/user/books?author="+author+"&page=1", http.StatusSeeOther)
		} else {
			if category == "" && author == "" && series != "" && name == "" {
				http.Redirect(rw, r, "/user/books?series="+series+"&page=1", http.StatusSeeOther)
			} else {
				if category == "" && author == "" && series == "" && name != "" {
					http.Redirect(rw, r, "/user/books?name="+name+"&page=1", http.StatusSeeOther)
				} else {
					a.GetBooksSearch(rw, "Должно быть заполнено только одно поле.")
					return
				}
			}
		}
	}
}

func (a app) GetBooksOpenID(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

	sp := filepath.Join("public", "html", "book-info.html")

	book, err := a.repo.GetBookById(a.ctx, p.ByName("id"))

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	tmpl, err := template.ParseFiles(sp, head)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "book-info", book)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}

func (a app) GetBooksReadID(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var con BookM
	sp := filepath.Join("public", "html", "book-read.html")

	book, err := a.repo.GetBookById(a.ctx, p.ByName("id"))

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	tmpl, err := template.ParseFiles(sp, head, header)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := os.ReadFile(book.Link)
	if err != nil {
		log.Fatal(err)
	}
	dec := charmap.Windows1251.NewDecoder()
	out, _ := dec.Bytes(data)
	con.Author = book.Author
	con.Name = book.Name
	con.Message = string(out)
	err = tmpl.ExecuteTemplate(rw, "book-read", con)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}

func (a app) AddNewBookPage(rw http.ResponseWriter, message string) {
	lp := filepath.Join("public", "html", "book.html")

	tmpl, err := template.ParseFiles(lp, head)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "book", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) AddNewBook(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	category := strings.TrimSpace(r.FormValue("category"))
	author := strings.TrimSpace(r.FormValue("author"))
	series := strings.TrimSpace(r.FormValue("series"))
	name := strings.TrimSpace(r.FormValue("name"))
	annotation := strings.TrimSpace(r.FormValue("annotation"))
	access := strings.TrimSpace(r.FormValue("access"))
	link := strings.TrimSpace(r.FormValue("link"))
	if category == "" || author == "" || series == "" || name == "" || annotation == "" || access == "" || link == "" {
		a.AddNewBookPage(rw, "Все поля должны быть заполнены")

	}

	err := a.repo.AddNewBook(a.ctx, category, author, series, name, annotation, link, access, time.Now())
	if err != nil {
		a.AddNewBookPage(rw, fmt.Sprintf("Ошибка создания книги: %v", err))
		return
	}
	http.Redirect(rw, r, "/admin/books", http.StatusSeeOther)
}

func (a app) DeleteBook(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.DeleteBookById(a.ctx, p.ByName("id"))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(rw, r, "/admin/books", http.StatusSeeOther)

}

func (a app) EditBookPage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sp := filepath.Join("public", "html", "bookedit.html")

	book, err := a.repo.GetBookById(a.ctx, p.ByName("id"))

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	tmpl, err := template.ParseFiles(sp, head, headera)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = tmpl.ExecuteTemplate(rw, "bookedit", book)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) EditBook(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	category := strings.TrimSpace(r.FormValue("category"))
	author := strings.TrimSpace(r.FormValue("author"))
	series := strings.TrimSpace(r.FormValue("series"))
	name := strings.TrimSpace(r.FormValue("name"))
	annotation := strings.TrimSpace(r.FormValue("annotation"))
	access := strings.TrimSpace(r.FormValue("access"))
	link := strings.TrimSpace(r.FormValue("link"))

	err := a.repo.PutBookById(a.ctx, p.ByName("id"), category, author, series, name, annotation, link, access, time.Now())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(rw, r, "/admin/books", http.StatusSeeOther)
}
func readCookie(name string, r *http.Request) (value string, err error) {
	if name == "" {
		return value, errors.New("you are trying to read empty cookie")
	}
	cookie, err := r.Cookie(name)
	if err != nil {
		return value, err
	}
	str := cookie.Value
	value, _ = url.QueryUnescape(str)
	return value, err
}

func NewApp(ctx context.Context, dbpool *pgxpool.Pool) *app {
	return &app{ctx, repository.NewRepository(dbpool), make(map[string]repository.User)}
}
