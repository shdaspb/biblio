package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

type ganr string

const (
	DETECTIVE  ganr = "Детективы, остросюжетная литература"
	CLASSIC    ganr = "Классика"
	ADVENTURES ganr = "Приключения, историческая литература"
	FANTASY    ganr = "Фантастика"
	HUMOR      ganr = "Юмор"
	KIND       ganr = "Детские"
	LOVE       ganr = "Любовно-слезоточивая литература"
	MODERN     ganr = "Современная литература"
)

type Book struct {
	Book_Id     uuid.UUID `json:"book_id" db:"book_id"`
	Category    ganr      `json:"category" db:"category"`
	Author      string    `json:"author" db:"author"`
	Series      string    `json:"series" db:"series"`
	Name        string    `json:"name" db:"name"`
	Annotation  string    `json:"annotation" db:"annotation"`
	Link        string    `json:"link" db:"link"`
	Access      string    `json:"access" db:"access"`
	Publication time.Time `json:"publication" db:"publication"`
}

type Page struct {
	Books      []Book
	Str        []int
	PageCount  int
	Number     int
	NextNumber int
	PrevNumber int
	PageUrl    string
}

type BookM struct {
	Book    Book
	Message string
}

func (r *Repository) AddNewBook(ctx context.Context, CategoryS, Author, Series, Name, Annotation, Link, Access string, Publication time.Time) (err error) {

	_, err = r.pool.Exec(ctx, `insert into books (category, author, series, name, annotation, link, access, publication) values ($1, $2,$3, $4, $5,$6,$7,$8)`, CategoryS, Author, Series, Name, Annotation, Link, Access, Publication)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) AllBook(ctx context.Context, pageNumber, pageSize int, s ...string) (page Page, err error) {
	var qwery string = ""
	var p Page
	if s[0] == "" && s[1] == "" && s[2] == "" && s[3] == "" && s[4] == "" {
		qwery = "select book_id, category, author, name from books order by category, author"
	}

	if s[0] != "" && s[1] == "" && s[2] == "" && s[3] == "" && s[4] == "" {
		qwery = "select book_id, category, author, name from books where category ilike '%" + s[0] + "%' order by category, author"
	}

	if s[0] == "" && s[1] != "" && s[2] == "" && s[3] == "" && s[4] == "" {
		qwery = "select book_id, category, author, name from books where author ilike '%" + s[1] + "%' order by category, author"
	}

	if s[0] == "" && s[1] == "" && s[2] != "" && s[3] == "" && s[4] == "" {
		qwery = "select book_id, category, author, name from books where series ilike '%" + s[2] + "%' order by category, author"
	}

	if s[0] == "" && s[1] == "" && s[2] == "" && s[3] != "" && s[4] == "" {
		qwery = "select book_id, category, author, name from books where name ilike '%" + s[3] + "%' order by category, author"
	}
	if s[0] == "" && s[1] == "" && s[2] == "" && s[3] == "" && s[4] != "" {
		qwery = "select book_id, category, author, name from books where link ilike '%" + s[4] + "%' order by category, author"
	}
	rows, err := r.pool.Query(ctx, qwery)
	if err != nil {
		err = fmt.Errorf("failed to query data: %w", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var b Book
		err = rows.Scan(&b.Book_Id, &b.Category, &b.Author, &b.Name)

		if err != nil {
			err = fmt.Errorf("failed to scan data: %w", err)
			return
		}
		p.Books = append(p.Books, b)
	}

	p.PageCount = int(math.Ceil(float64(len(p.Books)) / float64(pageSize)))

	for i := 1; i <= p.PageCount; i++ {
		p.Str = append(p.Str, i)
	}

	start := (pageNumber - 1) * pageSize
	end := pageNumber * pageSize

	if start >= len(p.Books) {
		return
	}

	if end > len(p.Books) {
		end = len(p.Books)
	}
	p.Books = p.Books[start:end]
	p.Number = pageNumber
	p.NextNumber = pageNumber + 1
	p.PrevNumber = pageNumber - 1
	page = p

	return
}

func (r *Repository) DeleteBookById(ctx context.Context, id string) (err error) {
	_, err = r.pool.Exec(ctx, `delete from books where book_id = $1`, id)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) GetBookById(ctx context.Context, id string) (b Book, err error) {
	rows := r.pool.QueryRow(ctx, `select * from books where book_id = $1`, id)

	err = rows.Scan(&b.Book_Id, &b.Category, &b.Author, &b.Series, &b.Name, &b.Annotation, &b.Link, &b.Access, &b.Publication)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return b, err
}

func (r *Repository) PutBookById(ctx context.Context, id, CategoryS, Author, Series, Name, Annotation, Link, Access string, Publication time.Time) (err error) {
	_, err = r.pool.Exec(ctx, `update books set category = $2, author = $3, series=$4, name=$5, annotation=$6, link=$7, access=$8, publication=$9 where book_id = $1`, id, CategoryS, Author, Series, Name, Annotation, Link, Access, Publication)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}
