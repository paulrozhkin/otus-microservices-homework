package entity

type User struct {
	Id        int64  `json:"id"`
	Username  string `json:"username" binding:"required,max=256"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
}
