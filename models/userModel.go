package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// this is basically a interface to connect with mongo db.
//beacuse mongo go doesn,t understand json and mongo need json.
//so it basically connect the two interfaces json and go.

type User struct{
	ID				primitive.ObjectID		`bson:"_id"`
	First_name		*string					`json:"first_name" validate:"required,min=2,max=100"`
	Last_name		*string					`json:"last_name" validate:"required,min=2,max=100"`
	Password		*string					`json:"Password" validate:"required,min=6"`
	Email			*string					`json:"email" validate:"email,required"`
	Phone			*string					`json:"phone" validate:"required"`
	Token			*string					`json:"token"`
	User_type		*string					`json:"user_type" validate:"required,eq=ADMIN|eq=USER"`//it is like enum validation in js, that only this particular type can access.
	Refresh_token	*string					`json:"refresh_token"`
	Created_at		time.Time				`json:"created_at"`
	Updated_at		time.Time				`json:"updated_at"`
	User_id			string					`json:"user_id"`
	IsVerified		bool					`json:"is_verified" bson:"is_verified"`
	VerifyToken		*string					`json:"verify_token" bson:"verify_token"`
	VerifyExpires	time.Time				`json:"verify_expires" bson:"verify_expires"`
}