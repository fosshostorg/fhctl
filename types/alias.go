package types

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type AliasItem struct {
	Type resourceType
	Id   string
}

type resourceType int

const (
	Project resourceType = iota
	Vm
	Proxy
	Pop
	Plan
)

// Search for project alias
func SearchProjectAlias(q string) (item AliasItem, err error) {
	var itemType resourceType

	id := viper.GetString(fmt.Sprintf("alias.%v.id", q))
	itemType = resourceType(viper.GetInt(fmt.Sprintf("alias.%v.id", q)))
	if id == "" {
		return AliasItem{}, errors.New("not found")
	}

	return AliasItem{Id: id, Type: itemType}, nil
}
