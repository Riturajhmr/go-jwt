package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)


// this file is made to check the authentication using uid, if caller is valid or not, to call this data.

func CheckUserType(c *gin.Context, role string) (err error){
	userType := c.GetString("user_type")
	err = nil
	if userType != role {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, userId string) (err error){
	userType := c.GetString("user_type")
	uid := c.GetString("uid")
	err= nil

	if userType == "USER" && uid != userId {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	err = CheckUserType(c, userType)
	return err
}