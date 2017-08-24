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
)

const (
	JSWKey = "muchWoW"
	PSWDKey = "suchAmaze"

)

type ToDoServer struct {
	db *gorm.DB
}

func NewToDoServer(db *gorm.DB) *ToDoServer {
	return &ToDoServer{db:db}
}

type ToDoItem struct {
	ID 			uint `json:"id",gorm:"auto_increment"`
	IsDone      bool `json:"done"`
	Description string `json:"description"`
}

type User struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

type NameToken struct {
	Username string `json:"username"`
	Token string `json:"token"`
}


func (Todoserver *ToDoServer) GetDatabase() *gorm.DB {
	return Todoserver.db
}

func (Todoserver *ToDoServer) handler (c echo.Context) error {
	var dat []ToDoItem
	Todoserver.db.Find(&dat)
	return c.JSON(http.StatusOK, dat)
}

func (Todoserver *ToDoServer) postHandler (c echo.Context) error {
	todo := ToDoItem{}
	c.Bind(&todo)
	Todoserver.db.Create(&todo)
	var dat []ToDoItem
	Todoserver.db.Find(&dat)
	return c.JSON(http.StatusOK, dat)
}

func (Todoserver *ToDoServer) deleteHandler (c echo.Context) error {
	id := c.Param("id")
	key, _ := strconv.Atoi(id)
	Todoserver.db.Delete(ToDoItem{}, "id = ?", key)
	var dat []ToDoItem
	Todoserver.db.Find(&dat)
	return c.JSON(http.StatusOK, dat)
}

func (Todoserver *ToDoServer) registerHandler (c echo.Context) error {

	// Get data from POST
	user := User{}
	c.Bind(&user)

	if Todoserver.db.Where("username = ?", user.Username).First(&User{}).RecordNotFound() {
		key, _ := scrypt.Key([]byte(user.Password), []byte(PSWDKey), 16384, 8, 1, 32)
		dbUser := User {user.Username, key}
		Todoserver.db.Create(&dbUser)
		return c.JSON(http.StatusOK, dbUser)
	}
	return c.JSON(http.StatusBadRequest, errors.New("Username already used!").Error())
}

func (Todoserver ToDoServer) loginHandler(c echo.Context) error {
	user := User{}
	c.Bind(&user)
	key, _ := scrypt.Key(user.Password, []byte(PSWDKey), 16384, 8, 1, 32)
	var dbUser User
	if Todoserver.db.Where("username = ? ", user.Username).First(&dbUser).RecordNotFound() {
		return c.JSON(http.StatusBadRequest, errors.New("Bad username or password!").Error())
	}
	if bytes.Compare(key, dbUser.Password) == 0 {
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": dbUser.Username,
			"password": dbUser.Password,
		})

		tokenString, _ := t.SignedString([]byte(JSWKey))
		Todoserver.db.Create(&NameToken{dbUser.Username, tokenString})
		fmt.Println(tokenString)

		return c.JSON(http.StatusOK, dbUser)
	} else {
		return c.JSON(http.StatusBadRequest, errors.New("Bad password!").Error())
	}
}

func TokenToStruct(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		return nil
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
	admin.Use()

	e.POST("/register", todoserver.registerHandler)
	e.POST("/login", todoserver.loginHandler)

	admin.GET("/", todoserver.handler)
	admin.DELETE("/:id", todoserver.deleteHandler)
	admin.POST("/", todoserver.postHandler)


	e.Logger.Fatal(e.Start(":8080"))
}

/*
		tokenString := c.Response().Header().Get("Authorization")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(SigningKey), nil
		})
		var user User
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			user = User { claims["username"].(string), claims["password"].(string)}
		} else {
			fmt.Println(err)
		}

		if err := Todoserver.db.Where("username = ? AND password = ?", user.Username, user.Password).First(User{}).Error; err != nil {
			return c.NoContent(http.StatusUnauthorized)
		}

return next(c)

*/

/*


	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SigningKey), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["username"], claims["password"])
	} else {
		fmt.Println(err)
	}


	return c.String(http.StatusOK, tokenString)

	*/