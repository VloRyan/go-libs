package jsonapi

import "github.com/vloryan/go-libs/httpx/router"

type ResourceHandler interface {
	RegisterRoutes(route router.RouteElement)
}

type ResourceIdentifierSource interface {
	GetIdentifier() *ResourceIdentifierObject
}
type ResourceIdentifierDestination interface {
	SetIdentifier(id *ResourceIdentifierObject)
}

type ResourceObjectMarshaller interface {
	MarshalResourceObject(*ResourceObject) error
}

type ResourceIdentifierObject struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	LID  string `json:"lid,omitempty"`
}

func (r *ResourceIdentifierObject) Equals(other *ResourceIdentifierObject) bool {
	if r == nil {
		return other == nil
	}
	if other == nil {
		return false
	}
	return r.ID == other.ID && r.Type == other.Type
}

type ResourceObject struct {
	ResourceIdentifierObject
	Attributes    map[string]any                 `json:"attributes,omitempty"`
	Relationships map[string]*RelationshipObject `json:"relationships,omitempty"`
	Links         map[string]any                 `json:"links,omitempty"`
	Meta          MetaData                       `json:"meta,omitempty"`
}

func (r *ResourceObject) LocalObjects() map[string]*ResourceObject {
	if r.Meta == nil {
		return nil
	}
	if v, ok := r.Meta["local-objects"]; ok {
		return v.(map[string]*ResourceObject)
	}
	return nil
}

func (r *ResourceObject) SetLocalObjects(m map[string]*ResourceObject) {
	if r.Meta == nil {
		r.Meta = MetaData{}
	}
	if m != nil {
		r.Meta["local-objects"] = m
	} else {
		delete(r.Meta, "local-objects")
	}
}

func (d *Document) NewResourceObject(id, typeStr string) *ResourceObject {
	return &ResourceObject{
		ResourceIdentifierObject: ResourceIdentifierObject{
			ID:   id,
			Type: typeStr,
		},
		Attributes:    make(map[string]any),
		Relationships: make(map[string]*RelationshipObject),
		Meta:          make(MetaData),
	}
}
