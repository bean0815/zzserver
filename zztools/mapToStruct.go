package zztools

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"
)

func FillStructSliceWithTag(mps []map[string]interface{}, obj interface{}) (objs []interface{}, err error) {
	if len(mps) == 0 {
		return nil, errors.New("len==0")
	}
	objs = make([]interface{}, len(mps))
	for index, mp := range mps {
		err := FillStructWithTag(mp, objs[index])
		return objs, err
	}
	return objs, nil
}

//用map填充结构_自定义struct字段标签
func FillStructWithTag(mp map[string]interface{}, obj interface{}) (err error) {
	elem := reflect.ValueOf(obj).Elem() //结构体值
	reftype := elem.Type()              //得到结构体类型
	fieldN := reftype.NumField()        //结构体字段数量
	if fieldN == 0 {
		return errors.New("没有字段")
	}

	for i := 0; i < reftype.NumField(); i++ {
		elemi := elem.Field(i)
		if !elem.IsValid() {
			log.Print("!elem.IsValid()")
			continue
		}
		if !elemi.CanSet() {
			log.Print("!elemi.CanSet()")
			continue
		}

		f := reftype.Field(i).Tag.Get("zz")
		// log.Println("tagz=>", f)

		if value, ok := mp[f]; ok {
			refval := reflect.ValueOf(value) //map 反射值
			// log.Println("map 反射值", refval)

			if elemi.Type() != refval.Type() {
				newRefVal, err := TypeConversion(fmt.Sprintf("%v", refval), elemi.Type().Name())
				if err != nil {
					log.Print("err=", err)
					return nil
				}
				refval = newRefVal
			}
			elemi.Set(refval)
		}
	}
	return nil
}

//用map填充结构
func FillStruct(data map[string]interface{}, obj interface{}) error {
	for k, v := range data {
		err := SetField(obj, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

//用map的值替换结构的值
func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()        //结构体属性值
	structFieldValue := structValue.FieldByName(name) //结构体单个属性值
	// log.Printf("field1=>%+v", field1)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type() //结构体的类型
	val := reflect.ValueOf(value)              //map值的反射值

	var err error
	if structFieldType != val.Type() {
		val, err = TypeConversion(fmt.Sprintf("%v", value), structFieldValue.Type().Name()) //类型转换
		if err != nil {
			return err
		}
	}

	structFieldValue.Set(val)
	return nil
}

//类型转换
func TypeConversion(value string, ntype string) (reflect.Value, error) {
	if ntype == "string" {
		return reflect.ValueOf(value), nil
	} else if ntype == "time.Time" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
		return reflect.ValueOf(t), err
	} else if ntype == "Time" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
		return reflect.ValueOf(t), err
	} else if ntype == "int" {
		i, err := strconv.Atoi(value)
		return reflect.ValueOf(i), err
	} else if ntype == "int8" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(int8(i)), err
	} else if ntype == "int32" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(int64(i)), err
	} else if ntype == "int64" {
		i, err := strconv.ParseInt(value, 10, 64)
		return reflect.ValueOf(i), err
	} else if ntype == "float32" {
		i, err := strconv.ParseFloat(value, 64)
		return reflect.ValueOf(float32(i)), err
	} else if ntype == "float64" {
		i, err := strconv.ParseFloat(value, 64)
		return reflect.ValueOf(i), err
	}

	//else if .......增加其他一些类型的转换

	return reflect.ValueOf(value), errors.New("未知的类型：" + ntype)
}
