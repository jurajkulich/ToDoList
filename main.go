package main

import (
	"github.com/labstack/echo"
	"net/http"
	"github.com/satori/go.uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"fmt"
)

type ToDoServer struct {
	TodoList ToDoList `json:"server"`
}

type ToDoList struct {
	Items []ToDoItem `json:"items"`
}

type ToDoItem struct {

	ID 			uint `json:"id",gorm:"auto_increment"`
	IsDone      bool `json:"done"`
	Description string `json:"description"`
}

func newTodoList() ToDoList {
	Todolist := ToDoList{}

	return Todolist
}

func (Todolist *ToDoList) AddItem(Todoitem ToDoItem) {

}

func (Todolist *ToDoList) GetItems() []ToDoItem {
	return Todolist.Items
}

func (Todolist *ToDoList) DeleteItem(key uuid.UUID) {
	// delete(Todolist.Items, key)
}

func (Todoserver *ToDoServer) GetList() ToDoList {
	return Todoserver.TodoList
}

func (Todoserver *ToDoServer) addServerItem(Item ToDoItem) {
	Todoserver.TodoList.AddItem(Item)
}

func (Todoserver *ToDoServer) GetServerItems() []ToDoItem {
	return Todoserver.TodoList.GetItems()
}

func (Todoserver *ToDoServer) handler(c echo.Context) error {
	return c.JSON(http.StatusOK, Todoserver.GetServerItems())
}

func (Todoserver *ToDoServer) postHandler(c echo.Context) error {
	item := ToDoItem{}
	c.Bind(&item)
	Todoserver.TodoList.AddItem(item)
	return c.JSON(http.StatusOK, Todoserver)
}

func (Todoserver *ToDoServer) deleteServerItem(key uuid.UUID) {
	Todoserver.TodoList.DeleteItem(key)
}

func (Todoserver *ToDoServer) deleteHandler(c echo.Context) error {
	id := c.Param("id")
	key, _ := uuid.FromString(id)
	Todoserver.deleteServerItem(key)
	return c.JSON(http.StatusOK, Todoserver)
}

func main() {
	db, err := gorm.Open("sqlite3", "/tmp/gorm.db")
	if err != nil {
		panic("Can't connect to database")
	}
	db.AutoMigrate(&ToDoItem{})
	defer db.Close()

	// db.Create(&ToDoItem{IsDone:false, Description:"buying pizza"})

	db.DropTable(ToDoItem{})

	todolist := newTodoList()
	Todoserver := ToDoServer{todolist}
	e := echo.New()
	e.GET("/", Todoserver.handler)
	e.DELETE("/:id", Todoserver.deleteHandler)
	e.POST("/", Todoserver.postHandler)
	e.Logger.Fatal(e.Start(":8080"))
}