package rconv2

import (
	"context"
	"errors"
	"strings"
)

// MapFilter A filter used in commands that return list of maps, e.g. Maps or MapRotation.
// The filter should return true, when the map should be included in the result set and false
// when the map should be skipped.
type MapFilter func(idx int, name string, result []string) bool

// GetAvailableMaps Returns a list of maps that are configured, and thus "available", and ready to be used in other commands.
// The list might be different from the various typed list elements that commands, such as ChangeMap, take as an argument.
// Whenever you want to issue a command that requires a map as its parameter, you _should_ check with GetAvailableMaps if the map
// is configured/available on the server. Alternatively, you may issue the command and respond to error messages if the map
// you chose is not available.
func (c *Connection) GetAvailableMaps(ctx context.Context, filters ...MapFilter) ([]string, error) {
	res, err := c.GetClientReferenceData(ctx, "AddMapToRotation")
	if err != nil {
		return nil, err
	}
	var param GetClientReferenceDataParameter
	for _, p := range res.Parameters {
		if p.Id == "MapName" {
			param = p
			break
		}
	}
	if param.Id != "MapName" {
		return nil, errors.New("could not find map name parameter")
	}
	maps := strings.Split(param.ValueMember, ",")
	return filter(maps, filters...), nil
}

func filter(maps []string, filters ...MapFilter) []string {
	var result []string
	for i, m := range maps {
		add := true
		for _, filter := range filters {
			if !filter(i, m, result) {
				add = false
			}
		}
		if add {
			result = append(result, m)
		}
	}
	return result
}
