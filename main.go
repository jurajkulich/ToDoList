package main

import (
	"github.com/labstack/echo"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"net/http"
	"strconv"
	"github.com/labstack/echo/middleware"
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
	Password string `json:"password"`
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
	user := User{}
	c.Bind(&user)
	Todoserver.db.Create(&user)
	return c.JSON(http.StatusOK, user)
}

func main() {
	database, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		panic("Can't connect to database")
	}
	database.AutoMigrate(&ToDoItem{})
	database.AutoMigrate(&User{})
	defer database.Close()

	todoserver := NewToDoServer(database)
	// todoserver.db.DropTable(ToDoItem{})
	// todoserver.db.DropTable(User{})
	e := echo.New()
	admin := e.Group("/admin")

	admin.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		var dbUser User
		if err := todoserver.db.Where("username = ? AND password = ?", username, password).First(&dbUser).Error; err != nil {
			return false, nil
		}
		return true, nil
	}))

	admin.GET("/", todoserver.handler)
	admin.DELETE("/:id", todoserver.deleteHandler)
	admin.POST("/", todoserver.postHandler)
	e.POST("/register", todoserver.registerHandler)

	e.Logger.Fatal(e.Start(":8080"))
}