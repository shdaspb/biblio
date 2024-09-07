package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type roles string

const (
	ADMIN roles = "ADMIN"
	USER  roles = "USER"
)

type User struct {
	User_Id        uuid.UUID `json:"user_id" db:"user_id"`
	Username       string    `json:"login" db:"usernamen"`
	Role           roles     `json:"role" db:"role"`
	FullName       string    `json:"full_name" db:"full_name"`
	HashedPassword string    `json:"hashed_password" db:"hashed_password"`
	Active         bool      `json:"activ" db:"active"`
}

func (r *Repository) Login(ctx context.Context, login, hashedPassword string) (u User, err error) {
	row := r.pool.QueryRow(ctx, `select user_id, username, role, full_name, active from users where username = $1 AND hashed_password = $2`, login, hashedPassword)

	err = row.Scan(&u.User_Id, &u.Username, &u.Role, &u.FullName, &u.Active)

	if err != nil {
		err = fmt.Errorf("failed to query data: %w", err)
		return
	}

	return
}

func (r *Repository) AddNewUser(ctx context.Context, username, full_name, hashedPassword string) (err error) {
	roles := USER
	active := true
	_, err = r.pool.Exec(ctx, `insert into users (username, role, full_name, hashed_password, active) values ($1, $2,$3, $4, $5)`, username, roles, full_name, hashedPassword, active)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) AllUser(ctx context.Context) (users []User, err error) {
	rows, err := r.pool.Query(ctx, `select user_id, username, role, full_name, active from users`)
	if err != nil {
		err = fmt.Errorf("failed to query data: %w", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		err = rows.Scan(&u.User_Id, &u.Username, &u.Role, &u.FullName, &u.Active)
		if err != nil {
			err = fmt.Errorf("failed to scan data: %w", err)
			return
		}
		users = append(users, u)
	}
	return
}

func (r *Repository) DeleteUserById(ctx context.Context, id string) (err error) {
	_, err = r.pool.Exec(ctx, `delete from users where user_id = $1`, id)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) GetUserById(ctx context.Context, id string) (u User, err error) {
	rows := r.pool.QueryRow(ctx, `select * from users where user_id = $1`, id)

	err = rows.Scan(&u.User_Id, &u.Username, &u.Role, &u.FullName, &u.HashedPassword, &u.Active)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return u, err
}

func (r *Repository) PutUserById(ctx context.Context, role, act, id string) (err error) {
	_, err = r.pool.Exec(ctx, `update users set role = $1, active = $2 where user_id = $3`, role, act, id)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}
