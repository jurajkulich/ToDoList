package main

import (
	"github.com/labstack/echo"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"fmt"
	"net/http"
	"strconv"
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

func main() {
	database, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		panic("Can't connect to database")
	}
	database.AutoMigrate(&ToDoItem{})
	defer database.Close()

	todoserver := NewToDoServer(database)

	var res []ToDoItem
	todoserver.db.Find(&res)
	fmt.Println(res)
	// todoserver.db.DropTable(ToDoItem{})

	e := echo.New()
	e.GET("/", todoserver.handler)
	e.DELETE("/:id", todoserver.deleteHandler)
	e.POST("/", todoserver.postHandler)
	e.Logger.Fatal(e.Start(":8080"))
}