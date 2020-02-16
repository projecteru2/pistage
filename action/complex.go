package action

import (
	"context"
	"path/filepath"

	"github.com/projecteru2/aa/config"
	"github.com/projecteru2/aa/errors"
	"github.com/projecteru2/aa/store"
	"github.com/projecteru2/aa/store/meta"
)

// Complex is combination of some actions,
// which are grouped to different paralled units,
// and all of them are shared same Parameters.
type Complex struct {
	Name   string     `json:"name"`
	Groups Groups     `json:"uses"`
	Params Parameters `json:"with,omitempty"`

	*meta.Ver `json:"-"`
}

// LoadComplex loads from the meta storage.
func LoadComplex(name string) (*Complex, error) {
	ctx, cancel := meta.Context(context.Background())
	defer cancel()

	c := NewComplex()
	c.Name = name

	var dict map[string]interface{}
	ver, err := store.Get(ctx, c.MetaKey(), &dict)
	if err != nil {
		return nil, errors.Trace(err)
	}

	c.SetVer(ver)

	if err := c.Parse(dict); err != nil {
		return nil, errors.Annotatef(err, "parse %s from meta failed", c.Name)
	}

	return c, meta.Load(c)
}

// NewComplex .
func NewComplex() *Complex {
	return &Complex{
		Groups: Groups{},
		Params: Parameters{},
		Ver:    meta.NewVer(),
	}
}

// Equal .
func (c *Complex) Equal(other *Complex) bool {
	switch {
	case c.Name != other.Name:
	case c.Groups == nil && other.Groups != nil:
	case c.Groups != nil && !c.Groups.Equal(other.Groups):
	case c.Params == nil && other.Params != nil:
	case c.Params != nil && !c.Params.Equal(other.Params):
	default:
		return true
	}
	return false
}

// Parse parses Complex fields' value from a raw dict.
func (c *Complex) Parse(dict map[string]interface{}) error {
	switch name, exists, err := mustString(dict, "name"); {
	case err != nil:
		return errors.Trace(err)
	case !exists:
		return errors.Annotatef(errors.ErrKeyNotExists, "name")
	case len(name) < 1:
		return errors.Annotatef(errors.ErrInvalidValue, "name is empty")
	default:
		c.Name = name
	}

	if raw, exists := dict[KeyParams]; exists {
		dict, ok := raw.(map[string]interface{})
		if !ok {
			return errors.Annotatef(errors.ErrInvalidType,
				"must be map[string]interface{} type, but %v", raw)
		}

		if err := c.Params.Parse(dict); err != nil {
			return errors.Trace(err)
		}
	}

	var err error
	if raw, exists := dict[KeyGroups]; exists {
		err = errors.Trace(c.Groups.Parse(raw))
	} else {
		err = errors.Trace(c.parseAction(dict))
	}
	if err != nil {
		return errors.Trace(err)
	}

	return c.Check()
}

// Check .
func (c *Complex) Check() error {
	if len(c.Name) < 1 {
		return errors.Annotatef(errors.ErrInvalidValue, "name is empty")
	}

	if err := c.Groups.Check(); err != nil {
		return errors.Trace(err)
	}

	if err := c.Params.Check(); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (c *Complex) parseAction(dict map[string]interface{}) error {
	atom := &Atom{}
	if err := atom.Parse(dict); err != nil {
		return errors.Trace(err)
	}

	groups := Groups{}
	if err := groups.Set(MainGroupName, Actions{atom}); err != nil {
		return errors.Trace(err)
	}

	c.Groups = groups

	return nil
}

// Save .
func (c *Complex) Save(ctx context.Context) error {
	return meta.Save(meta.Resources{c})
}

// MetaKey .
func (c *Complex) MetaKey() string {
	return filepath.Join(config.Conf.EtcdPrefix, "complex", c.Name)
}
