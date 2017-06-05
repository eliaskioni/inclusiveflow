package events

// TypeUpdateContact is the type of our update contact event
const TypeUpdateContact string = "update_contact"

// UpdateContact events are created when a contact's built in field is updated.
//
// ```
//   {
//    "step_uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
//    "type": "update_contact",
//    "created_on": "2006-01-02T15:04:05Z",
//    "field_name": "Language",
//    "value": "eng"
//   }
// ```
//
// @event update_contact
type UpdateContactEvent struct {
	BaseEvent
	FieldName string `json:"field_name"  validate:"required"`
	Value     string `json:"value"       validate:"required"`
}

// NewUpdateContact returns a new save to contact event
func NewUpdateContact(name string, value string) *UpdateContactEvent {
	return &UpdateContactEvent{
		FieldName: name,
		Value:     value,
	}
}

// Type returns the type of this event
func (e *UpdateContactEvent) Type() string { return TypeUpdateContact }
