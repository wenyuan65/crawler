package main

import (
	"fmt"
	"time"

	"github.com/wy/crawler/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Food struct {
	Id         int
	Name       string
	Price      float64
	TypeId     int
	CreateTime *time.Time `gorm:"column:createtime"`
}

func (food Food) TableName() string {
	return "food"
}

func main() {
	dataSourceName := config.GetValue("jdbc.dataSourceName")
	db, _ := connect(dataSourceName)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	var foods []Food
	db.Find(&foods)
	fmt.Println(foods)

	// db.Create(foods)
	// db.Delete()
	// db.Update()
	// db.UpdateColumns()
}

func connect(dataSourceUrl string) (db *gorm.DB, err error) {
	// driverName := config.GetValue("jdbc.driverName")

	db, err2 := gorm.Open(mysql.Open(dataSourceUrl), &gorm.Config{})
	if err2 != nil {
		fmt.Println(err2)
	}

	return db, nil
}
