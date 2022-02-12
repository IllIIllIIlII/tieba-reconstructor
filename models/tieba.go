package models

type TiebaThread struct {
	FloorNumber uint
	Author      User
	Id          string
	Title       string
	Floors      []Floor
}
type Floor struct {
	FloorNumber int
	Content     string
	Author      User
	Id          string
	Comments    []Comment
}
type User struct {
	Id          int    `json:"user_id"` // tieba user id
	DisplayName string `json:"un"`
	Url         string
}
type Comment struct {
	Content string
	Authro  User
}
