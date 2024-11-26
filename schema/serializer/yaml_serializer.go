package serializer

import (
	"context"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"
	"gorm.io/gorm/schema"
)

type YamlSerializer struct{}

// 反序列化
func (ser YamlSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue any) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var bs []byte
		switch v := dbValue.(type) {
		case []byte:
			bs = v
		case string:
			bs = []byte(v)
		default:
			return fmt.Errorf("failed to unmarshal YAML value: %#v", dbValue)
		}

		if len(bs) > 0 {
			err = yaml.Unmarshal(bs, fieldValue.Interface())
		}
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return
}

// 序列化
func (ser YamlSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue any) (val any, err error) {
	bs, errM := yaml.Marshal(fieldValue)
	fmt.Println(field.TagSettings)
	return string(bs), errM
}
