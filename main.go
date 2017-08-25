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
	JWTKey = "muchWoW"
	PSWDKey = "suchAmaze"
)

type ToDoServer struct {
	db *gorm.DB
}

func NewToDoServer(db *gorm.DB) *ToDoServer {
	return &ToDoServer{db:db}
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


func (Todoserver *ToDoServer) GetDatabase() *gorm.DB {
	return Todoserver.db
}

func (Todoserver *ToDoServer) handler (c echo.Context) error {
	user := &User{}
	userName := c.Get("context")
	err := Todoserver.db.First(user, "username = ?", userName).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}
	Todoserver.db.Model(&user).Related(&user.Todoitem)
	return c.JSON(http.StatusOK, user)
}

func (Todoserver *ToDoServer) postHandler (c echo.Context) error {
	todo := ToDoItem{}
	c.Bind(&todo)
	if todo.Name == "" {
		return c.JSON(http.StatusBadRequest, "Bad input. You have to add \"name\"")
	}
	user := &User{}
	userName := c.Get("context")
	err := Todoserver.db.First(user, "username = ?", userName).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}
	todo.UserID = user.ID
	Todoserver.db.Create(&todo)
	return c.JSON(http.StatusOK, todo)
}

func (Todoserver *ToDoServer) deleteHandler (c echo.Context) error {
	user := &User{}
	userName := c.Get("context")
	id := c.Param("id")
	key, _ := strconv.Atoi(id)

	todo := ToDoItem{}
	err := Todoserver.db.First(user, "username = ?", userName).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}
	todo.UserID = user.ID

	if Todoserver.db.Delete(&todo, "id = ?", key).RecordNotFound() {
		return c.JSON(http.StatusBadRequest, "No item found")
	}
	return c.JSON(http.StatusOK, "Deleted")
}

func (Todoserver *ToDoServer) registerHandler (c echo.Context) error {

	// Get data from POST
	user := User{}
	c.Bind(&user)
	if user.Username == ""  {
		return c.JSON(http.StatusBadRequest, errors.New("Bad input!").Error())
	}

	if Todoserver.db.Where("username = ?", user.Username).First(&User{}).RecordNotFound() {
		key, _ := scrypt.Key([]byte(user.Password), []byte(PSWDKey), 16384, 8, 1, 32)
		dbUser := User {Username: user.Username, Password:key}
		Todoserver.db.Create(&dbUser)
		return c.JSON(http.StatusOK, dbUser)
	}
	return c.JSON(http.StatusBadRequest, errors.New("Username already used!").Error())
}

func (Todoserver ToDoServer) loginHandler(c echo.Context) error {

	// get user from POST
	user := User{}
	c.Bind(&user)
	// crypt user password
	key, _ := scrypt.Key(user.Password, []byte(PSWDKey), 16384, 8, 1, 32)

	var dbUser User
	if Todoserver.db.Where("username = ? ", user.Username).First(&dbUser).RecordNotFound() {
		return c.JSON(http.StatusBadRequest, errors.New("Bad username or password!").Error())
	}
	if bytes.Compare(key, dbUser.Password) == 0 {
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"exp": time.Now().Add(time.Minute * 60).Unix(),
			"username": dbUser.Username,
		})

		tokenString, _ := t.SignedString([]byte(JWTKey))
		nameToken := &NameToken{dbUser.Username, tokenString}
		Todoserver.db.Create(nameToken)

		return c.JSON(http.StatusOK, nameToken)
	} else {
		return c.JSON(http.StatusBadRequest, errors.New("Bad password!").Error())
	}
}

func (Todoserver *ToDoServer) updateHandler(c echo.Context) error {
	userName := c.Get("context")
	id := c.Param("id")
	todo := ToDoItem{}
	c.Bind(&todo)
	user := User{}
	err := Todoserver.db.First(&user, "username = ?", userName).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}

	todo.UserID = user.ID
	key, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}
	todo.ID = uint(key)
	oldTodo := &ToDoItem{}
	err = Todoserver.db.First(oldTodo, "id = ?", id ).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}

	err = Todoserver.db.Save(&todo).Error
	if err != nil {
		return c.JSON(http.StatusNoContent, err)
	}
	return c.JSON(http.StatusOK, oldTodo)
}

func (Todoserver *ToDoServer) TokenToStruct(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString  := c.Request().Header.Get("authorization")

		if Todoserver.db.Where("token = ? ", tokenString).First(&NameToken{}).RecordNotFound() {
			return c.NoContent(http.StatusUnauthorized)
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTKey), nil
		})

		claims,ok := token.Claims.(jwt.MapClaims)

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
	database, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		panic("Can't connect to database")
	}

	database.AutoMigrate(&ToDoItem{})
	database.AutoMigrate(&User{})
	database.AutoMigrate(&NameToken{})
	defer database.Close()

	todoserver := NewToDoServer(database)

	// todoserver.db.DropTable(ToDoItem{})
	// todoserver.db.DropTable(User{})
	// todoserver.db.DropTable(NameToken{})

	e := echo.New()
	admin := e.Group("/admin")
	admin.Use(todoserver.TokenToStruct)

	e.POST("/register", todoserver.registerHandler)
	e.POST("/login", todoserver.loginHandler)

	admin.GET("/", todoserver.handler)
	admin.DELETE("/:id", todoserver.deleteHandler)
	admin.POST("/", todoserver.postHandler)
	admin.POST("/:id", todoserver.updateHandler)

	e.Logger.Fatal(e.Start(":8080"))
}
