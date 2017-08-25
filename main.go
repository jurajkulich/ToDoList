package main

import (
	"github.com/labstack/echo"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"net/http"
	"strconv"
	"github.com/dgrijalva/jwt-go"
	"fmt"
	"errors"
	"golang.org/x/crypto/scrypt"
	"bytes"
	"time"
)


const (
	// JWT key
	JWTKey = "muchWoW"
	// Hash password key
	PSWDKey = "suchAmaze"
)

type ToDoServer struct {
	db *gorm.DB
}

type ToDoItem struct {
	gorm.Model
	Name string	`json:"name"`
	Description string `json:"description"`
	IsDone      bool `json:"done"`
	UserID uint
}

type User struct {
	gorm.Model
	Username string `json:"username"`
	Password []byte `json:"password"`
	Todoitem []ToDoItem
}

type NameToken struct {
	Username string `json:"username"`
	Token string `json:"token"`
}

// Initialize database in ToDoServer
func NewToDoServer(db *gorm.DB) *ToDoServer {
	return &ToDoServer{db:db}
}

// Shows user's info and his ToDoItems
func (Todoserver *ToDoServer) getHandler (c echo.Context) error {
	user := &User{}
	// Get username from MiddleWare TokenToStruct func
	userName := c.Get("context")
	// Check user in database
	err := Todoserver.db.First(user, "username = ?", userName).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}
	// Give user Has Many relationship with ToDoItem
	Todoserver.db.Model(&user).Related(&user.Todoitem)
	return c.JSON(http.StatusOK, user)
}

// Add ToDoItem to user's database
func (Todoserver *ToDoServer) postHandler (c echo.Context) error {
	todo := ToDoItem{}
	c.Bind(&todo)
	// ToDoItem name can't be empty
	if todo.Name == "" {
		return c.JSON(http.StatusBadRequest, "Bad input. You have to add \"name\"")
	}
	user := &User{}
	// Check user in database
	userName := c.Get("context")
	err := Todoserver.db.First(user, "username = ?", userName).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}

	todo.UserID = user.ID

	// Add new user to databse
	Todoserver.db.Create(&todo)
	return c.JSON(http.StatusOK, todo)
}

// Delete ToDoItem by ID
func (Todoserver *ToDoServer) deleteHandler (c echo.Context) error {
	user := &User{}
	// Get name from MiddleWare TokenToStruct func
	userName := c.Get("context")
	// Get ToDoItem ID from URL
	id := c.Param("id")
	// Convert string ID to int id
	key,err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}

	todo := ToDoItem{}

	// Check user in database
	err = Todoserver.db.First(user, "username = ?", userName).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}

	todo.UserID = user.ID

	// Return error when ToDoItem with this ID was not found
	if Todoserver.db.Delete(&todo, "id = ?", key).RecordNotFound() {
		return c.JSON(http.StatusBadRequest, "No item found")
	}
	return c.JSON(http.StatusOK, "Deleted")
}

// User register by username and password
func (Todoserver *ToDoServer) registerHandler (c echo.Context) error {

	// Get data from POST
	user := User{}
	c.Bind(&user)
	// Check if is username empty(error)
	if user.Username == ""  {
		return c.JSON(http.StatusBadRequest, errors.New("Bad input!").Error())
	}
	// When checked name isn't in database create new user
	if Todoserver.db.Where("username = ?", user.Username).First(&User{}).RecordNotFound() {
		// Hash password
		key,err := scrypt.Key([]byte(user.Password), []byte(PSWDKey), 16384, 8, 1, 32)
		if err != nil {
			return c.JSON(http.StatusNoContent, err)
		}

		// Make user with username and hashed password
		dbUser := User {Username: user.Username, Password:key}

		// Add new user to database
		Todoserver.db.Create(&dbUser)

		return c.JSON(http.StatusOK, dbUser)
	}
	return c.JSON(http.StatusBadRequest, errors.New("Username already used!").Error())
}

// User login, make token, add it to database, return token
func (Todoserver ToDoServer) loginHandler(c echo.Context) error {

	// get user from POST
	user := User{}
	c.Bind(&user)
	// crypt user password
	key, err := scrypt.Key(user.Password, []byte(PSWDKey), 16384, 8, 1, 32)
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}

	// Check if user exists in database
	var dbUser User
	if Todoserver.db.Where("username = ? ", user.Username).First(&dbUser).RecordNotFound() {
		return c.JSON(http.StatusBadRequest, errors.New("Bad username").Error())
	}

	// Check if user's password equals password in database
	if bytes.Compare(key, dbUser.Password) == 0 {

		// Create new user's token with claims "exp" and "username" and signing method HS256
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"exp": time.Now().Add(time.Minute * 60).Unix(),
			"username": dbUser.Username,
		})

		// Encode token as string using the secret
		tokenString, err := t.SignedString([]byte(JWTKey))
		if err != nil {
			return c.JSON(http.StatusNoContent, err)
		}

		// Add new token to database with username
		nameToken := &NameToken{dbUser.Username, tokenString}
		Todoserver.db.Create(nameToken)

		return c.JSON(http.StatusOK, nameToken)
	} else {
		return c.JSON(http.StatusBadRequest, errors.New("Bad password!").Error())
	}
}

// Update ToDoItem by ID
func (Todoserver *ToDoServer) updateHandler(c echo.Context) error {

	// Get username from MiddleWare TokenToStruct func
	userName := c.Get("context")

	// Get ToDoItem ID from URL
	id := c.Param("id")

	// Add data from POST to ToDoItem
	todo := ToDoItem{}
	c.Bind(&todo)

	// Check user in database
	user := User{}
	err := Todoserver.db.First(&user, "username = ?", userName).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}

	todo.UserID = user.ID

	// Parse ID from URL from string to uint
	key, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}
	todo.ID = uint(key)

	// Save new ToDoItem to database
	err = Todoserver.db.Save(&todo).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}
	return c.JSON(http.StatusOK, todo)
}

/// MiddleWare func checks user's token validity
func (Todoserver *ToDoServer) TokenToStruct(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		// Get user's token from header
		tokenString  := c.Request().Header.Get("authorization")

		// Check if is token in database
		if Todoserver.db.Where("token = ? ", tokenString).First(&NameToken{}).RecordNotFound() {
			return c.NoContent(http.StatusUnauthorized)
		}

		// Get token key for validating
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTKey), nil
		})

		// Get token claims
		claims,ok := token.Claims.(jwt.MapClaims)

		// Check token validity and set username in echo to other handler functions
		if err == nil && token.Valid && ok {
			username := claims["username"]
			c.Set("context",  username)

			fmt.Println("Token is valid")
			return next(c)
		} else {
			fmt.Printf("Bad token")
			return c.NoContent(http.StatusUnauthorized)
		}
	}
}

func main() {
	// Open database
	database, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		panic("Can't connect to database")
	}

	database.AutoMigrate(&ToDoItem{})
	database.AutoMigrate(&User{})
	database.AutoMigrate(&NameToken{})
	defer database.Close()

	// Initialize database in ToDoServer struct
	todoserver := NewToDoServer(database)

	// todoserver.db.DropTable(ToDoItem{})
	// todoserver.db.DropTable(User{})
	// todoserver.db.DropTable(NameToken{})

	e := echo.New()
	admin := e.Group("/admin")

	// MiddleWare func checks user's token validity
	admin.Use(todoserver.TokenToStruct)

	// User register by username and password
	e.POST("/register", todoserver.registerHandler)
	// User login, make token, add it to database, return token
	e.POST("/login", todoserver.loginHandler)

	// Shows user's info and his ToDoItems
	admin.GET("/", todoserver.getHandler)
	// Delete ToDoItem by ID
	admin.DELETE("/:id", todoserver.deleteHandler)
	// Add ToDoItem to user's database
	admin.POST("/", todoserver.postHandler)
	// Update ToDoItem by ID
	admin.POST("/:id", todoserver.updateHandler)

	e.Logger.Fatal(e.Start(":8080"))
}

/*
func (Todoserver ToDoServer) getUserbyName(username string) (User, error) {
	user := User{}
	err := Todoserver.db.First(&user, "username = ?", username).Error
	if err != nil {
		return user, err
	}
	return user, err
}
 */