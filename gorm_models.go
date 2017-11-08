package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

//FieldTypes FieldTypes
var FieldTypes map[string]string

func main() {
	FieldTypes = map[string]string{
		"bigint":    "int64",
		"int":       "int",
		"tinyint":   "int",
		"smallint":  "int",
		"char":      "string",
		"varchar":   "string",
		"blob":      "[]byte",
		"date":      "time.Time",
		"datetime":  "time.Time",
		"timestamp": "time.Time",
		"decimal":   "float64",
		"bit":       "uint64",
		"enum":      "string",
		"text":      "string",
		"set":       "string",
		"float":     "float32",
		"double":    "float64",
	}
	Init()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//Init init
func Init() {
	db, err := sql.Open("mysql", "root:antonio@tcp(localhost:3306)/wms?charset=utf8&parseTime=true")

	modelres, err := db.Prepare("SHOW TABLES")
	check(err)

	modelqry, err := modelres.Query()
	check(err)

	var modelName string
	for modelqry.Next() {
		modelqry.Scan(&modelName)

		res, err := db.Prepare("DESCRIBE " + modelName)

		check(err)

		qry, err := res.Query()
		check(err)
		// Field | Type | Null | Key | Default | Extra
		var (
			Field   string
			Type    string
			Null    string
			Key     string
			Default string
			Extra   string
		)
		f, err := os.Create("models/" + modelName + ".go")
		check(err)

		modelnames := strings.Split(modelName, "_")
		modelName = ""
		for _, n := range modelnames {
			modelName += strings.Title(n)
		}

		f.WriteString(`package models
			
				import "time"
		
				// ` + modelName + ` Model
				type ` + modelName + ` struct {
				`)

		for qry.Next() {
			qry.Scan(&Field, &Type, &Null, &Key, &Default, &Extra)
			Title := strings.Title(Field)
			i := strings.Index(Type, "(")
			if i != -1 {
				Type = Type[0:i]
			}

			fmt.Printf("Field=%s, Type=%s, Null=%s, Key=%s, Default=%s, Extra=%s \n", Field, Type, Null, Key, Default, Extra)
			var name string
			names := strings.Split(Title, "_")

			for _, n := range names {
				if len(names) == 1 {
					if len(n) <= 3 {
						name = strings.ToUpper(n)
					} else {
						name = strings.Title(n)
					}
				} else {
					var word string
					if strings.ToUpper(n) == "ID" {
						word = strings.ToUpper(n)
					} else {
						word = strings.Title(n)
					}

					name += word
				}
			}

			tp := FieldTypes[Type]
			sql := "`json:\"" + Field + "\" gorm:\"column:" + Field + ";"
			if Null == "NO" {
				sql += "NOT NULL;"
			}
			if Key == "PRI" {
				sql += "PRIMARY KEY;"

				if strings.Contains(Extra, "auto_increment") {
					sql += "AUTO_INCREMENT;"
				}
			}

			sql += "\"`"
			line := fmt.Sprintf("  %-10s\t%-10s\t%-20s", name, tp, sql)
			f.WriteString(line + "\n")
			Field, Type, Null, Key, Default, Extra = "", "", "", "", "", ""
		}

		f.WriteString("}")
		res.Close()
		qry.Close()
		f.Close()

		modelName = ""
	}

	modelres.Close()
	modelqry.Close()

	println("All Done!! Have Fun!!")
	defer db.Close()

}
