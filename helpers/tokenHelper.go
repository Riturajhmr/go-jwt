package helper

import (
	"context"
	"fmt"
	"jwtauth/database"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// jwt token basically uses a hashing mechanism, by taking the details,
// you give it, and generate a token.
// and in website JWT.IO you can decode the token and take all values back.
// and jwt also generate a secret key which keep all your secrets and generate the
// token for you, it basically used to store the private information.
type SignedDetails struct{
	Email 		string
	First_name 	string
	Last_name 	string
	Uid 		string
	User_type	string
	jwt.StandardClaims 
}


var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, userType string, uid string) (signedToken string, signedRefreshToken string, err error){
	claims := &SignedDetails{
		Email : email,
		First_name: firstName,
		Last_name: lastName,
		Uid : uid,
		User_type: userType,
		//for how much duration the token will last.
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	//it is used to re assign the token after expiry.
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
			//The ExpiresAt field ensures that the token has a limited lifespan,
			// enhancing security by forcing users to re-authenticate after the token expires.
		},
	}

	token ,_ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	// jwt.SigningMethodHS256, it is a algorithm to create a encrypted token for you.
	/*SignedString([]byte(SECRET_KEY)):

This method signs the token using the HMAC-SHA256 algorithm and the provided secret key.
SECRET_KEY is your application's private secret key (a string). It should be securely stored and kept confidential.
It converts the secret key into a byte slice ([]byte) since the signing method requires binary data.

token:
	The resulting token is a string in JWT format, which consists of three parts:

	Header: Encoded metadata about the token (e.g., algorithm, type).
	Payload: Encoded claims (user data, expiration, etc.).
	Signature: A cryptographic signature created using the secret key to verify the token's authenticity.
These parts are separated by dots (.)*/

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return 
	}

	return token, refreshToken, err
}


func ValidateToken(signedToken string) (claims *SignedDetails, msg string){
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token)(interface{}, error){
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok:= token.Claims.(*SignedDetails)
	if !ok{
		
		fmt.Println("the token is invalid")
		msg = err.Error()
		return 
	}

	if claims.ExpiresAt < time.Now().Local().Unix(){
		fmt.Println("token is expired")
		msg = err.Error()
		return
	}
	return claims, msg
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string){
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj primitive.D
	/*primitive.D is a type in the MongoDB Go driver that represents an ordered BSON document.
	It is essentially a slice of key-value pairs and is used when you need to preserve the order of fields in a document,
	which is important in certain MongoDB operations (e.g., $set, $push, etc.).*/


	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	// bson.E: This represents a BSON element with a key-value pair.

	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})


	//upsert is set to true, meaning if no matching document is found,
	//MongoDB will create a new document using the specified update object (updateObj).
	upsert := true
	filter := bson.M{"user_id":userId}

	//This encapsulates(or contains) the upsert behavior for the update operation.
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{Key: "$set", Value: updateObj},
		},
		&opt,
	)

	defer cancel()

	if err!=nil{
		log.Panic(err)
		return
	}
}