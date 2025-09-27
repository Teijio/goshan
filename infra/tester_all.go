// package main

// import (
// 	"context"
// 	"fmt"
// 	"time"
// )

// type DB struct {
// }

// type User struct {
// 	Name string
// }

// func (d *DB) SelectUser(ctx context.Context, email string) (User, error) {
// 	timer := time.NewTimer(1 * time.Second)
// 	select {
// 	case <-timer.C:
// 		return User{Name: "Gosha"}, nil
// 	case <-ctx.Done():
// 		return User{}, fmt.Errorf("context canceled")
// 	}
// }

// type Handler struct {
// 	db *DB
// }

// type Request struct {
// 	Email string
// }

// type Response struct {
// 	User User
// }

// func (h *Handler) HandleAPI(ctx context.Context, req Request) (Response, error) {
// 	u, err := h.db.SelectUser(ctx, req.Email)
// 	if err != nil {
// 		return Response{}, err
// 	}

// 	return Response{User: u}, nil
// }

// func main() {
// 	db := DB{}
// 	handler := Handler{db: &db}
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1500)
// 	defer cancel()

// 	req := Request{Email: "test@yandex.ru"}
// 	resp, err := handler.HandleAPI(ctx, req)
// 	fmt.Println(resp, err)
// }
