package gorm

import (
	"errors"
	"fmt"
	"reflect"
)

type Association struct {
	Scope      *Scope
	PrimaryKey interface{}
	Column     string
	Error      error
	Field      *Field
}

func (association *Association) setErr(err error) *Association {
	if err != nil {
		association.Error = err
	}
	return association
}

func (association *Association) Find(value interface{}) *Association {
	association.Scope.related(value, association.Column)
	return association.setErr(association.Scope.db.Error)
}

func (association *Association) Append(values ...interface{}) *Association {
	scope := association.Scope
	field := association.Field

	for _, value := range values {
		reflectvalue := reflect.Indirect(reflect.ValueOf(value))
		if reflectvalue.Kind() == reflect.Struct {
			field.Set(reflect.Append(field.Field, reflectvalue))
		} else if reflectvalue.Kind() == reflect.Slice {
			field.Set(reflect.AppendSlice(field.Field, reflectvalue))
		} else {
			association.setErr(errors.New("invalid association type"))
		}
	}
	scope.callCallbacks(scope.db.parent.callback.updates)
	return association.setErr(scope.db.Error)
}

func (association *Association) getPrimaryKeys(values ...interface{}) []interface{} {
	primaryKeys := []interface{}{}
	scope := association.Scope

	for _, value := range values {
		reflectValue := reflect.Indirect(reflect.ValueOf(value))
		if reflectValue.Kind() == reflect.Slice {
			for i := 0; i < reflectValue.Len(); i++ {
				if primaryField := scope.New(reflectValue.Index(i).Interface()).PrimaryKeyField(); !primaryField.IsBlank {
					primaryKeys = append(primaryKeys, primaryField.Field.Interface())
				}
			}
		} else if reflectValue.Kind() == reflect.Struct {
			if primaryField := scope.New(value).PrimaryKeyField(); !primaryField.IsBlank {
				primaryKeys = append(primaryKeys, primaryField.Field.Interface())
			}
		}
	}
	return primaryKeys
}

func (association *Association) Delete(values ...interface{}) *Association {
	primaryKeys := association.getPrimaryKeys(values...)

	if len(primaryKeys) == 0 {
		association.setErr(errors.New("no primary key found"))
	} else {
		relationship := association.Field.Relationship
		// many to many
		if relationship.Kind == "many_to_many" {
			whereSql := fmt.Sprintf("%v.%v = ? AND %v.%v IN (?)",
				relationship.JoinTable, association.Scope.Quote(relationship.ForeignDBName),
				relationship.JoinTable, association.Scope.Quote(relationship.AssociationForeignDBName))

			if err := association.Scope.DB().Table(relationship.JoinTable).
				Where(whereSql, association.PrimaryKey, primaryKeys).Delete("").Error; err == nil {
				leftValues := reflect.Zero(association.Field.Field.Type())
				for i := 0; i < association.Field.Field.Len(); i++ {
					value := association.Field.Field.Index(i)
					if primaryField := association.Scope.New(value.Interface()).PrimaryKeyField(); primaryField != nil {
						var included = false
						for _, primaryKey := range primaryKeys {
							if equalAsString(primaryKey, primaryField.Field.Interface()) {
								included = true
							}
						}
						if !included {
							leftValues = reflect.Append(leftValues, value)
						}
					}
				}
				association.Field.Set(leftValues)
			}
		} else {
			association.setErr(errors.New("delete only support many to many"))
		}
	}
	return association
}

func (association *Association) Replace(values ...interface{}) *Association {
	relationship := association.Field.Relationship
	scope := association.Scope
	if relationship.Kind == "many_to_many" {
		field := association.Field.Field

		oldPrimaryKeys := association.getPrimaryKeys(field.Interface())
		association.Field.Set(reflect.Zero(association.Field.Field.Type()))
		association.Append(values...)
		newPrimaryKeys := association.getPrimaryKeys(field.Interface())

		var addedPrimaryKeys = []interface{}{}
		for _, newKey := range newPrimaryKeys {
			hasEqual := false
			for _, oldKey := range oldPrimaryKeys {
				if reflect.DeepEqual(newKey, oldKey) {
					hasEqual = true
					break
				}
			}
			if !hasEqual {
				addedPrimaryKeys = append(addedPrimaryKeys, newKey)
			}
		}
		for _, primaryKey := range association.getPrimaryKeys(values...) {
			addedPrimaryKeys = append(addedPrimaryKeys, primaryKey)
		}

		whereSql := fmt.Sprintf("%v.%v = ? AND %v.%v NOT IN (?)",
			relationship.JoinTable, association.Scope.Quote(relationship.ForeignDBName),
			relationship.JoinTable, association.Scope.Quote(relationship.AssociationForeignDBName))

		scope.DB().Table(relationship.JoinTable).Where(whereSql, association.PrimaryKey, addedPrimaryKeys).Delete("")
	} else {
		association.setErr(errors.New("replace only support many to many"))
	}
	return association
}

func (association *Association) Clear() *Association {
	relationship := association.Field.Relationship
	scope := association.Scope
	if relationship.Kind == "many_to_many" {
		whereSql := fmt.Sprintf("%v.%v = ?", relationship.JoinTable, scope.Quote(relationship.ForeignDBName))
		if err := scope.DB().Table(relationship.JoinTable).Where(whereSql, association.PrimaryKey).Delete("").Error; err == nil {
			association.Field.Set(reflect.Zero(association.Field.Field.Type()))
		} else {
			association.setErr(err)
		}
	} else {
		association.setErr(errors.New("clear only support many to many"))
	}
	return association
}

func (association *Association) Count() int {
	count := -1
	relationship := association.Field.Relationship
	scope := association.Scope
	newScope := scope.New(association.Field.Field.Interface())

	if relationship.Kind == "many_to_many" {
		scope.DB().Table(relationship.JoinTable).
			Select("COUNT(DISTINCT ?)", relationship.AssociationForeignDBName).
			Where(relationship.ForeignDBName+" = ?", association.PrimaryKey).Row().Scan(&count)
	} else if relationship.Kind == "has_many" || relationship.Kind == "has_one" {
		whereSql := fmt.Sprintf("%v.%v = ?", newScope.QuotedTableName(), newScope.Quote(relationship.ForeignDBName))
		countScope := scope.DB().Table(newScope.TableName()).Where(whereSql, association.PrimaryKey)
		if relationship.PolymorphicType != "" {
			countScope = countScope.Where(fmt.Sprintf("%v.%v = ?", newScope.QuotedTableName(), newScope.Quote(relationship.PolymorphicDBName)), scope.TableName())
		}
		countScope.Count(&count)
	} else if relationship.Kind == "belongs_to" {
		if v, ok := scope.FieldByName(association.Column); ok {
			whereSql := fmt.Sprintf("%v.%v = ?", newScope.QuotedTableName(), newScope.Quote(relationship.ForeignDBName))
			scope.DB().Table(newScope.TableName()).Where(whereSql, v).Count(&count)
		}
	}

	return count
}
