package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type EmailDTO struct {
	// 邮箱
	Email string
}

type LikeDTO struct {
	ID uint
}

type IdDTO struct {
	ID uint
}

type IDflv struct {
	ID     uint
	Flvkey string
}

type ObjectIdDTO struct {
	ID primitive.ObjectID
}
