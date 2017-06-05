package flows

import (
	"encoding/json"
	"fmt"

	"github.com/nyaruka/goflow/utils"
)

type Contact struct {
	uuid     ContactUUID
	name     string
	language utils.Language
	urns     URNList
	groups   GroupList
	fields   *Fields
}

func (c *Contact) SetLanguage(lang utils.Language) { c.language = lang }
func (c *Contact) Language() utils.Language        { return c.language }

func (c *Contact) SetName(name string) { c.name = name }
func (c *Contact) Name() string        { return c.name }

func (c *Contact) URNs() URNList     { return c.urns }
func (c *Contact) UUID() ContactUUID { return c.uuid }

func (c *Contact) Groups() GroupList { return GroupList(c.groups) }
func (c *Contact) AddGroup(uuid GroupUUID, name string) {
	c.groups = append(c.groups, &Group{uuid, name})
}
func (c *Contact) RemoveGroup(uuid GroupUUID) bool {
	for i := range c.groups {
		if c.groups[i].uuid == uuid {
			c.groups = append(c.groups[:i], c.groups[i+1:]...)
			return true
		}
	}
	return false
}

func (c *Contact) Fields() *Fields { return c.fields }

func (c *Contact) Resolve(key string) interface{} {
	switch key {

	case "name":
		return c.name

	case "uuid":
		return c.uuid

	case "urns":
		return c.urns

	case "language":
		return c.language

	case "groups":
		return GroupList(c.groups)

	case "fields":
		return c.fields

	case "urn":
		return c.urns
	}

	return fmt.Errorf("No field '%s' on contact", key)
}

func (c *Contact) Default() interface{} {
	return c
}

type ContactReference struct {
	UUID ContactUUID `json:"uuid"    validate:"required,uuid4"`
	Name string      `json:"name"`
}

//------------------------------------------------------------------------------------------
// JSON Encoding / Decoding
//------------------------------------------------------------------------------------------

// ReadContact decodes a contact from the passed in JSON
func ReadContact(data json.RawMessage) (*Contact, error) {
	contact := &Contact{}
	err := json.Unmarshal(data, contact)
	if err == nil {
		// err = run.Validate()
	}
	return contact, err
}

type contactEnvelope struct {
	UUID     ContactUUID    `json:"uuid"`
	Name     string         `json:"name"`
	Language utils.Language `json:"language"`
	URNs     URNList        `json:"urns"`
	Groups   GroupList      `json:"groups"`
	Fields   *Fields        `json:"fields,omitempty"`
}

func (c *Contact) UnmarshalJSON(data []byte) error {
	var ce contactEnvelope
	var err error

	err = json.Unmarshal(data, &ce)
	if err != nil {
		return err
	}

	c.name = ce.Name
	c.uuid = ce.UUID
	c.language = ce.Language

	if ce.URNs == nil {
		c.urns = make(URNList, 0)
	} else {
		c.urns = ce.URNs
	}

	if ce.Groups == nil {
		c.groups = make(GroupList, 0)
	} else {
		c.groups = ce.Groups
	}

	if ce.Fields == nil {
		c.fields = NewFields()
	} else {
		c.fields = ce.Fields
	}

	return err
}

func (c *Contact) MarshalJSON() ([]byte, error) {
	var ce contactEnvelope

	ce.Name = c.name
	ce.UUID = c.uuid
	ce.Language = c.language
	ce.URNs = c.urns
	ce.Groups = c.groups
	ce.Fields = c.fields

	return json.Marshal(ce)
}