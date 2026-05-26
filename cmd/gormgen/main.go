package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	dsn := flag.String("dsn", "", "consult[https://gorm.io/docs/connecting_to_the_database.html]")
	tables := flag.String("tables", "", "enter the required data table or leave it blank")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("dsn cannot be empty, please provide a valid dsn value.")
	}

	db, err := gorm.Open(mysql.Open(*dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalf("failed to connect database: %s", err.Error())
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory:%s", err.Error())
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:          wd + "/internal/repository/mysql/dao",
		ModelPkgPath:     wd + "/internal/repository/mysql/model",
		Mode:             gen.WithoutContext | gen.WithDefaultQuery,
		FieldNullable:    true,
		FieldWithTypeTag: true,
	})
	g.WithImportPkgPath("github.com/shopspring/decimal")
	g.UseDB(db)

	applyModel := func(tableName string) interface{} {
		return g.GenerateModel(tableName,
			gen.FieldTypeReg("^(total_lessons|used_lessons|remain_lessons|total_amount|paid_amount|low_lesson_threshold|amount)$", "decimal.Decimal"),
			gen.FieldGenTypeReg("^(total_lessons|used_lessons|remain_lessons|total_amount|paid_amount|low_lesson_threshold|amount)$", "Float64"),
		)
	}

	tableMaps := strings.Split(*tables, ",")
	if len(tableMaps) == 0 || *tables == "" {
		tableNames, err := db.Migrator().GetTables()
		if err != nil {
			log.Fatalf("failed to list tables: %s", err.Error())
		}
		models := make([]interface{}, 0, len(tableNames))
		for _, tableName := range tableNames {
			models = append(models, applyModel(tableName))
		}
		g.ApplyBasic(models...)
	} else {
		models := make([]interface{}, 0, len(tableMaps))
		for _, tableName := range tableMaps {
			name := strings.TrimSpace(tableName)
			if name == "" {
				continue
			}
			models = append(models, applyModel(name))
		}
		g.ApplyBasic(models...)
	}

	g.Execute()
}
