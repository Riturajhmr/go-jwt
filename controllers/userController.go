package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	//it is a go library to validate the structs and fields.

	"jwtauth/database"
	helper "jwtauth/helpers"
	"jwtauth/models"
	"jwtauth/services"

	"golang.org/x/crypto/bcrypt"

	// it is use to securely store and validate the password.

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//to call the instance of collection to fetch the information
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

//To validate struct fields easily.
//To ensure that incoming data meets the expected format or constraints (e.g., email, required, length).
var validate = validator.New()

// in the database you can't store the password as it is,
// you have to hash it before storing, beacuse if not then any one,
// who has access to the database can get your password.
func HashPassword(password string) string{
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err!=nil{
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string)(bool, string){

	/*It compares a hashed password (typically stored securely in a database) with a plain-text password,
	 by rehashing the plain-text password and checking if the hashes match.*/
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err!= nil {
		msg = "email of password is incorrect"
		check=false
	}
	return check, msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		// Check if email already exists
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email already exists"})
			return
		}

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for the email"})
			return
		}

		// Check if phone already exists
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for the phone number"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this phone number already exists"})
			return
		}

		// Generate verification token
		verifyToken := services.GenerateVerificationToken()
		
		// Send verification email first
		emailService := services.NewEmailService()
		err = emailService.SendVerificationEmail(*user.Email, verifyToken)
		if err != nil {
			log.Printf("Failed to send verification email: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
			return
		}

		// Store verification data in temporary collection
		verificationData := bson.M{
			"email": *user.Email,
			"first_name": *user.First_name,
			"last_name": *user.Last_name,
			"password": HashPassword(*user.Password),
			"phone": *user.Phone,
			"user_type": *user.User_type,
			"verify_token": verifyToken,
			"verify_expires": services.GetVerificationExpiryTime(),
			"created_at": time.Now(),
		}

		// Store in temporary collection
		_, err = database.OpenCollection(database.Client, "pending_verifications").InsertOne(ctx, verificationData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store verification data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Verification email sent. Please verify your email to complete registration.",
		})
	}
}

func VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "verification token is required"})
			return
		}

		// Find verification data
		var verificationData bson.M
		err := database.OpenCollection(database.Client, "pending_verifications").FindOne(ctx, bson.M{"verify_token": token}).Decode(&verificationData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid verification token"})
			return
		}

		// Check if token has expired
		verifyExpires := verificationData["verify_expires"].(time.Time)
		if time.Now().After(verifyExpires) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "verification token has expired"})
			return
		}

		// Create new user
		user := models.User{
			ID: primitive.NewObjectID(),
			First_name: verificationData["first_name"].(*string),
			Last_name: verificationData["last_name"].(*string),
			Password: verificationData["password"].(*string),
			Email: verificationData["email"].(*string),
			Phone: verificationData["phone"].(*string),
			User_type: verificationData["user_type"].(*string),
			Created_at: time.Now(),
			Updated_at: time.Now(),
			User_id: primitive.NewObjectID().Hex(),
			IsVerified: true,
		}

		// Generate tokens
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		// Insert user into main collection
		_, err = userCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Delete verification data
		_, err = database.OpenCollection(database.Client, "pending_verifications").DeleteOne(ctx, bson.M{"verify_token": token})
		if err != nil {
			log.Printf("Failed to delete verification data: %v", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Email verified successfully. You can now login.",
			"user": user,
		})
	}
}

func Login() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return 
		}


		// by using the email, we store the user information, in foundUser struct
		err := userCollection.FindOne(ctx, bson.M{"email":user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"email or password is incorrect"})
			return
		}

		// and by using the verify password function we verfiy that the user password, will match with foundUser or not.
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true{
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"user not found"})
		}
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		err = userCollection.FindOne(ctx, bson.M{"user_id":foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc{
	return func(c *gin.Context){
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage <1{
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 !=nil || page<1{
			page = 1
		}

		

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}}, 
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, 
			//this $push to the root is done to access the data, if we not we only see the count, but not the data.
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}

			//it is basically to control the data rendering, which will be shown.
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data",  recordPerPage}}}},}}}
result,err := userCollection.Aggregate(ctx, mongo.Pipeline{
	matchStage, groupStage, projectStage})
defer cancel()
if err!=nil{
	c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
}
var allusers []bson.M
if err = result.All(ctx, &allusers); err!=nil{
	log.Fatal(err)
}
c.JSON(http.StatusOK, allusers[0])}}

// gin gives access to its own handler function.
func GetUser() gin.HandlerFunc{
	return func(c *gin.Context){
		userId := c.Param("user_id")

		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id":userId}).Decode(&user)
		// we use decode function beacuse go does not understand the json format.
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}