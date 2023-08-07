// nolint
package subspace

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/maticnetwork/heimdall/helper/structs"
	"github.com/mitchellh/mapstructure"
)

var (
	ErrDecode = errors.New("faield to decode")
)

func overwriteable(to, from *structs.Field) bool {
	if to.Kind() != from.Kind() {
		return false
	}

	if !from.IsExported() {
		return false
	}

	if from.Kind() == reflect.Ptr {
		if from.IsZero() {
			return false
		}
	}

	return true
}

// RecursiveMergeOverwrite will use ‘from‘ to overwrite ’to‘ in ’dst‘ if possible.
func RecursiveMergeOverwrite(from, to, dst interface{}) error {
	toMap := structs.Map(to)
	toStruct := structs.New(to)
	fromMap := structs.Map(from)
	fromStruct := structs.New(from)

	for k, fromValue := range fromMap {
		// mismatch field will be filtered
		_, ok := toMap[k]
		if !ok {
			return fmt.Errorf("Target key not found: %s.\n", k)
		}

		toField := toStruct.Field(k)
		fromField := fromStruct.Field(k)
		switch toField.Kind() {
		case reflect.Map:
			fromData := reflect.ValueOf(fromField.Value())
			fromKeys := fromData.MapKeys()
			toData := reflect.ValueOf(toField.Value())
			toKeys := toData.MapKeys()

			// tokenizer key to string to recognize same field,
			// and record original (key, value) with Pair
			// later the Pair will be write back to field
			type Pair struct {
				Key   interface{}
				Value interface{}
			}
			toDataCache := make(map[string]Pair)

			for _, k := range toKeys {
				key := k.Interface()
				value := toData.MapIndex(k).Interface()

				toDataCache[k.String()] = Pair{
					Key:   key,
					Value: value,
				}
			}
			for _, k := range fromKeys {
				val, ok := toDataCache[k.String()]
				// if fromData has different key, then add it to toDataCache derectly
				// else recursive merge the value field corresponding to the key
				if !ok {
					value := fromData.MapIndex(k)
					toDataCache[k.String()] = Pair{
						Key:   k.Interface(),
						Value: value.Interface(),
					}
				} else {
					fromValue := fromData.MapIndex(k).Interface()
					toValue := reflect.ValueOf(val.Value).Interface()

					if reflect.TypeOf(toValue).Kind() == reflect.Struct {
						// if map's value is a struct, then merge the value recursively
						err := RecursiveMergeOverwrite(fromValue, toValue, &toValue)
						if err != nil {
							return err
						}
					} else {
						// if map's value is not a struct, then merge the value
						innerData := reflect.ValueOf(&toValue).Elem()
						innerData.Set(reflect.ValueOf(fromValue))
					}

					val.Value = toValue
					toDataCache[k.String()] = val
				}
			}

			// write back to field
			newInstance := reflect.ValueOf(fromValue)
			if newInstance.IsNil() {
				fromValue = toMap[k]
				newInstance = reflect.ValueOf(fromValue)
			}

			for _, v := range toDataCache {
				key := reflect.ValueOf(v.Key)
				value := reflect.ValueOf(v.Value)
				newInstance.SetMapIndex(key, value)
			}
			toMap[k] = fromValue

		default:
			if overwriteable(toField, fromField) {
				toMap[k] = fromValue
			}
		}
	}

	if err := mapstructure.Decode(toMap, dst); err != nil {
		return ErrDecode
	}

	return nil
}
