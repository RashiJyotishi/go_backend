package models
import "go.mongodb.org/mongo-driver/bson/primitive"

type SignupRequest struct {
    Username        string `json:"username"`
    Email           string `json:"email"`
    Password        string `json:"password"`
    ConfirmPassword string `json:"confirm_password"`
}

type User struct {
    ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Username string             `json:"username" bson:"username"`
    Email    string             `json:"email" bson:"email"`
    Password string             `json:"password" bson:"password"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}