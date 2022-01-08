package models

type TiebaThread struct {
	FloorNumber uint
	Author      User
	Id          string
	Title       string
	Floors      []Floor
}
type Floor struct {
	Number   int //floor number
	Content  string
	Author   User
	Id       string
	Comments []Comment
}
type User struct {
	Id          int // tieba user id
	DisplayName string
	Url         string
}
type Comment struct {
	Content string
	Authro  User
}
