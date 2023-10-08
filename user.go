package main

type User struct {
	ID       uint   `gorm:"column:user_id;autoIncrement;primaryKey;"`
	Username string `gorm:"type:varchar(16);column:username;not null;"`
	Password string `gorm:"type:varchar(20);column:password;not null;"`
	Tokens   uint   `gorm:"type:integer;column:token;default:1000;not null;"`
}

func (User) TableName() string {
	return "user"
}
