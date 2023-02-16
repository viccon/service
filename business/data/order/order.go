// Package order provides support for describing the ordering of data.
package order

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ardanlabs/service/business/sys/validate"
)

// Individual directions in the system.
var (
	ASC  = Direction{"ASC"}
	DESC = Direction{"DESC"}
)

// Set of known directions.
var directions = map[string]Direction{
	ASC.name:  ASC,
	DESC.name: DESC,
}

// Direction defines an order direction.
type Direction struct {
	name string
}

// parseDirection converts a string to a type Direction.
func parseDirection(value string) (Direction, error) {
	direction, exists := directions[value]
	if !exists {
		return Direction{}, errors.New("invalid direction")
	}

	return direction, nil
}

// =============================================================================

// Field represents a field of database being managed.
type Field struct {
	name    string
	storage string
}

// NewField constructs a new field for the system.
func NewField(name string) Field {
	return Field{
		name: name,
	}
}

// AddStorageField constructs a Field value and checks for potential sql
// injection issues. If there is an error it will panic.
func (f *Field) AddStorageField(name string) Field {
	f.storage = name

	return *f
}

// =============================================================================

// FieldSet maintains a set of fields that belong to an entity.
type FieldSet struct {
	fields map[string]Field
}

// NewFieldSet takes a comma delimited set of fields to add to the set.
func NewFieldSet(fields ...Field) FieldSet {
	m := make(map[string]Field)

	for _, field := range fields {
		m[field.name] = field
	}

	return FieldSet{
		fields: m,
	}
}

// parseField takes a field by string and validates it belongs to the set.
// Then returns that field in its proper type.
func (fs FieldSet) parseField(field string) (Field, error) {
	f, exists := fs.fields[field]
	if !exists {
		return Field{}, fmt.Errorf("field %q not found", field)
	}

	return f, nil
}

// =============================================================================

// By represents a field used to order by and direction.
type By struct {
	field     Field
	direction Direction
}

// NewBy constructs a new By value with no checks.
func NewBy(field Field, direction Direction) By {
	by := By{
		field:     field,
		direction: direction,
	}

	return by
}

// Clause returns a sql string with the ordering information.
func (b By) Clause() (string, error) {
	return b.field.storage + " " + b.direction.name, nil
}

// =============================================================================

// Parse constructs an order.By value by parsing a string in the form
// of "field,direction" from the request.
func Parse(r *http.Request, orderingFields FieldSet, defaultOrder By) (By, error) {
	v := r.URL.Query().Get("orderBy")

	if v == "" {
		return defaultOrder, nil
	}

	orderParts := strings.Split(v, ",")

	var by By
	switch len(orderParts) {
	case 1:
		field, err := orderingFields.parseField(strings.Trim(orderParts[0], " "))
		if err != nil {
			return By{}, validate.NewFieldsError(v, errors.New("parsing fields"))
		}

		by = NewBy(field, ASC)

	case 2:
		field, err := orderingFields.parseField(strings.Trim(orderParts[0], " "))
		if err != nil {
			return By{}, validate.NewFieldsError(v, errors.New("parsing fields"))
		}

		dir, err := parseDirection(strings.Trim(orderParts[1], " "))
		if err != nil {
			return By{}, validate.NewFieldsError(v, errors.New("parsing direction"))
		}

		by = NewBy(field, dir)

	default:
		return By{}, validate.NewFieldsError(v, errors.New("unknown order field"))
	}

	return by, nil
}
