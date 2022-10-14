package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wy/crawler/config"
)

type users struct {
	Id   int    `db:"id"`
	name string `db:"name"`
}

func main() {
	driverName := config.GetValue("jdbc.driverName")
	dataSourceName := config.GetValue("jdbc.dataSourceName")

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		fmt.Println(err)
		return
	}

	// query
	rows, err2 := db.Query("select * from users")
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user users
		err = rows.Scan(&user.Id, &user.name)
		fmt.Println(user)
	}

	r, err3 := db.Exec("insert into users(id,name)value(?,?)", 5, "老王")
	if err3 != nil {
		fmt.Println(err3)
		return
	}

	lastInsertId, _ := r.LastInsertId()
	affected, _ := r.RowsAffected()
	fmt.Printf("id:%d, %d", lastInsertId, affected)

}
