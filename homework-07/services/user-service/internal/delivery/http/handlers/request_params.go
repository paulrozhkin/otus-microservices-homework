package handlers

type RequestParams struct {
	Id int64 `uri:"id" binding:"required"`
}
