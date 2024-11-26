package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"gitlab.miliantech.com/risk/base/risk_common/utils"
)

var CodeFormatTool string = "goimports"
var ModelDir string = "models"

type Gen struct {
	ModelName string
	TableName string
	FileName  string
	Fields    map[string]*Field

	OrderedFields []*Field
}

type Field struct {
	Type string
	Name string
	Tag  string
}

func (g *Gen) Parse() {
	flag.Parse()

	arg := strings.TrimSpace(flag.Arg(0))
	g.FileName = arg + ".go"
	g.TableName = arg
	modelName := titleHandler(arg)
	g.ModelName = modelName

	for i := 1; i < len(flag.Args()); i++ {
		items := strings.Split(flag.Arg(i), ":")
		if len(items) < 2 { // 参数错误
			continue
		}

		fieldName, typeName := items[0], items[1]
		field := &Field{}
		field.Name = titleHandler(strings.TrimSpace(fieldName))
		field.Type = strings.TrimSpace(typeName)
		if len(items) >= 3 {
			field.Tag = strings.TrimSpace(items[2])
		}

		if _, load := g.Fields[field.Name]; !load {
			g.Fields[field.Name] = field
			g.OrderedFields = append(g.OrderedFields, field)
		} else {
			log.Println(fieldName + " duplicate")
		}
	}
}

func (g *Gen) Generate() (err error) {
	// generate code
	code, _ := g.genCode()
	// generate file
	file, errFile := g.genFile()
	if errFile != nil {
		return errFile
	}
	defer file.Close()

	// write code into file
	_, err = file.WriteString(code)
	if err != nil {
		return err
	}
	// code format
	return exec.Command(CodeFormatTool, "-w", g.getFileName()).Run()
}

func (g *Gen) genFile() (file *os.File, err error) {

	err = os.MkdirAll(ModelDir, os.ModePerm)
	if err != nil {
		log.Println(fmt.Sprintf("create dir %s:", ModelDir), err)
		return nil, err
	}

	fileName := g.getFileName()
	_, errFile := os.Stat(fileName)
	if os.IsExist(errFile) {
		log.Println("create file:", errFile)
		return nil, errFile
	}

	file, err = os.Create(fileName)
	if err != nil {
		log.Println("create file:", err)
		return nil, err
	}
	return
}

func (g *Gen) getFileName() string {
	return ModelDir + "/" + g.FileName
}

func (g *Gen) genCode() (ctt string, err error) {
	bs := bytes.NewBuffer(nil)

	bs.WriteString("package " + strings.ToLower(ModelDir))
	bs.WriteString("\n")

	// import
	bs.WriteString("import (")
	bs.WriteString("\n")
	bs.WriteString(StringWarp("context"))
	bs.WriteString("\n")
	for _, item := range []string{
		"github.com/jinzhu/gorm",
		"gitlab.miliantech.com/infrastructure/log",
		"go.uber.org/zap",
	} {
		bs.WriteString(StringWarp(item))
		bs.WriteString("\n")
	}
	for _, field := range g.Fields {
		if utils.ArrayContains([]string{"time", "time.Time"}, field.Type) {
			bs.WriteString(StringWarp("time"))
			bs.WriteString("\n")
		}
	}

	bs.WriteString(")")
	bs.WriteString("\n\n")

	// model
	bs.WriteString("type " + g.ModelName + " struct {")
	// bs.WriteString("\n")
	// bs.WriteString("Id int64")
	bs.WriteString("\n")
	for _, of := range g.OrderedFields {
		field := g.Fields[of.Name]
		var fieldType string = field.Type
		if utils.ArrayContains([]string{"time", "time.Time"}, field.Type) {
			fieldType = "time.Time"
		}
		bs.WriteString(field.Name + " " + fieldType + " " + field.Tag + "\n")
	}
	// bs.WriteString("CreatedAt time.Time")
	// bs.WriteString("\n")
	// bs.WriteString("UpdatedAt time.Time")
	bs.WriteString("\n")
	bs.WriteString("}")
	bs.WriteString("\n\n")

	// TableName
	bs.WriteString("func (record *" + g.ModelName + ") TableName() string {")
	bs.WriteString(" return \"" + g.TableName + "\" ")
	bs.WriteString("}")
	bs.WriteString("\n\n")

	/*
		func CreateXxx(ctx context.Context, client *gorm.DB, record *Xxx) error {
			err := client.Create(record).Error
			if err != nil {
				log.Info(ctx, "risk_common.CreateXxx", zap.Error(err))
			}
			return nil
		}
	*/
	bs.WriteString("func Create" + g.ModelName + "(ctx context.Context, client *gorm.DB, record *" + g.ModelName + ") error {")
	bs.WriteString("\n")
	bs.WriteString("err := client.Create(record).Error")
	bs.WriteString("\n")
	bs.WriteString("if err != nil {")
	bs.WriteString("\n")
	bs.WriteString("log.Info(ctx, " + StringWarp("risk_common.Create"+g.ModelName) + ", zap.Error(err))")
	bs.WriteString("\n")
	bs.WriteString("}")
	bs.WriteString("\n")
	bs.WriteString("return nil\n}\n")
	ctt = bs.String()
	return
}

var titleHandler func(string) string = utils.CamelString

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_ = ctx

	g := &Gen{Fields: map[string]*Field{}}
	g.Parse()
	if err := g.Generate(); err != nil {
		log.Println("gen fail with error: ", err)
	} else {
		log.Println("gen success")
	}
}

func StringWarp(str string) string {
	return fmt.Sprintf("\"%s\"", str)
}
