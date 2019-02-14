package mobile

// To build an Android Archive:
//
// gomobile bind -target android -javapkg=com.nyaruka.goflow -o mobile/goflow.aar github.com/nyaruka/goflow/mobile
//
// ... except gomobile doesn't yet support gomodules (https://github.com/golang/go/issues/27234). So you need to recreate
// this as a non-module go project first, i.e.
//
// mkdir -p $GOPATH/src/github.com/nyaruka/goflow
// rsync -a . $GOPATH/src/github.com/nyaruka/goflow
// cd $GOPATH/src/github.com/nyaruka/goflow
// GO111MODULE=on go mod vendor
// GO111MODULE=off go get golang.org/x/mobile/cmd/gomobile
// $GOPATH/bin/gomobile init
// GO111MODULE=off gomobile bind -target android -javapkg=com.nyaruka.goflow -o mobile/goflow.aar github.com/nyaruka/goflow/mobile

import (
	"encoding/json"
	"time"

	"github.com/nyaruka/gocommon/urns"
	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/assets/static"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/definition"
	"github.com/nyaruka/goflow/flows/engine"
	"github.com/nyaruka/goflow/flows/resumes"
	"github.com/nyaruka/goflow/flows/triggers"
	"github.com/nyaruka/goflow/flows/waits"
	"github.com/nyaruka/goflow/utils"

	"github.com/Masterminds/semver"
)

// IsSpecVersionSupported returns whether the given flow spec version is supported
func IsSpecVersionSupported(ver string) bool {
	v, err := semver.NewVersion(ver)
	if err != nil {
		return false
	}
	return definition.IsSpecVersionSupported(v)
}

// Environment defines the environment for expression evaluation etc
type Environment struct {
	target utils.Environment
}

// NewEnvironment creates a new environment.
func NewEnvironment(dateFormat string, timeFormat string, timezone string, defaultLanguage string, allowedLanguages *StringSlice, defaultCountry string, redactionPolicy string) (*Environment, error) {
	tz, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	langs := make([]utils.Language, allowedLanguages.Length())
	for l := 0; l < allowedLanguages.Length(); l++ {
		langs[l] = utils.Language(allowedLanguages.Get(l))
	}

	return &Environment{
		target: utils.NewEnvironmentBuilder().
			WithDateFormat(utils.DateFormat(dateFormat)).
			WithTimeFormat(utils.TimeFormat(timeFormat)).
			WithTimezone(tz).
			WithDefaultLanguage(utils.Language(defaultLanguage)).
			WithAllowedLanguages(langs).
			WithDefaultCountry(utils.Country(defaultCountry)).
			WithRedactionPolicy(utils.RedactionPolicy(redactionPolicy)).
			Build(),
	}, nil
}

// AssetsSource is a static asset source
type AssetsSource struct {
	target *static.StaticSource
}

// NewAssetsSource creates a new static asset source
func NewAssetsSource(src string) (*AssetsSource, error) {
	s, err := static.NewSource(json.RawMessage(src))
	if err != nil {
		return nil, err
	}
	return &AssetsSource{target: s}, nil
}

// SessionAssets provides optimized access to assets
type SessionAssets struct {
	target flows.SessionAssets
}

// NewSessionAssets creates a new session assets
func NewSessionAssets(source *AssetsSource) (*SessionAssets, error) {
	s, err := engine.NewSessionAssets(source.target)
	if err != nil {
		return nil, err
	}
	return &SessionAssets{target: s}, nil
}

// Contact represents a person who is interacting with a flow
type Contact struct {
	target *flows.Contact
}

// NewEmptyContact creates a new contact
func NewEmptyContact(sa *SessionAssets) *Contact {
	return &Contact{
		target: flows.NewEmptyContact(sa.target, "", utils.NilLanguage, nil),
	}
}

// MsgIn is an incoming message
type MsgIn struct {
	target *flows.MsgIn
}

// NewMsgIn creates a new incoming message
func NewMsgIn(uuid string, text string, attachments *StringSlice) *MsgIn {
	var convertedAttachments []flows.Attachment
	if attachments != nil {
		convertedAttachments = make([]flows.Attachment, attachments.Length())
		for a := 0; a < attachments.Length(); a++ {
			convertedAttachments[a] = flows.Attachment(attachments.Get(a))
		}
	}

	return &MsgIn{
		target: flows.NewMsgIn(flows.MsgUUID(uuid), urns.NilURN, nil, text, convertedAttachments),
	}
}

func (m *MsgIn) Text() string {
	return m.target.Text()
}

func (m *MsgIn) Attachments() *StringSlice {
	attachments := NewStringSlice(len(m.target.Attachments()))
	for attachment := range m.target.Attachments() {
		attachments.Add(string(attachment))
	}
	return attachments
}

// FlowReference is a reference to a flow
type FlowReference struct {
	uuid string
	name string
}

// NewFlowReference creates a new flow reference
func NewFlowReference(uuid string, name string) *FlowReference {
	return &FlowReference{uuid: uuid, name: name}
}

// Trigger represents something which can initiate a session
type Trigger struct {
	target flows.Trigger
}

// NewManualTrigger creates a new manual trigger
func NewManualTrigger(environment *Environment, contact *Contact, flow *FlowReference) *Trigger {
	flowRef := assets.NewFlowReference(assets.FlowUUID(flow.uuid), flow.name)
	return &Trigger{
		target: triggers.NewManualTrigger(environment.target, flowRef, contact.target, nil),
	}
}

// Resume represents something which can resume a session
type Resume struct {
	target flows.Resume
}

// NewMsgResume creates a new message resume
func NewMsgResume(environment *Environment, contact *Contact, msg *MsgIn) *Resume {
	var e utils.Environment
	if environment != nil {
		e = environment.target
	}
	var c *flows.Contact
	if contact != nil {
		c = contact.target
	}

	return &Resume{
		target: resumes.NewMsgResume(e, c, msg.target),
	}
}

type Event struct {
	type_   string
	payload string
}

func (e *Event) Type() string {
	return e.type_
}

func (e *Event) Payload() string {
	return e.payload
}

type Modifier struct {
	type_   string
	payload string
}

func (m *Modifier) Type() string {
	return m.type_
}

func (m *Modifier) Payload() string {
	return m.payload
}

// Sprint is an interaction with the engine - i.e. a start or resume of a session
type Sprint struct {
	target flows.Sprint
}

// Modifiers returns the modifiers created during this sprint
func (s *Sprint) Modifiers() *ModifierSlice {
	mods := NewModifierSlice(len(s.target.Modifiers()))
	for _, mod := range s.target.Modifiers() {
		marshaled, _ := json.Marshal(mod)
		mods.Add(&Modifier{type_: mod.Type(), payload: string(marshaled)})
	}
	return mods
}

// Events returns the events created during this sprint
func (s *Sprint) Events() *EventSlice {
	events := NewEventSlice(len(s.target.Events()))
	for _, event := range s.target.Events() {
		marshaled, _ := json.Marshal(event)
		events.Add(&Event{type_: event.Type(), payload: string(marshaled)})
	}
	return events
}

// Session represents a session with the flow engine
type Session struct {
	target flows.Session
}

// Status returns the status of this session
func (s *Session) Status() string {
	return string(s.target.Status())
}

// Assets returns the assets associated with this session
func (s *Session) Assets() *SessionAssets {
	return &SessionAssets{target: s.target.Assets()}
}

// Start starts this session using the given trigger
func (s *Session) Start(trigger *Trigger) (*Sprint, error) {
	sprint, err := s.target.Start(trigger.target)
	if err != nil {
		return nil, err
	}
	return &Sprint{target: sprint}, nil
}

// Resume resumes this session
func (s *Session) Resume(resume *Resume) (*Sprint, error) {
	sprint, err := s.target.Resume(resume.target)
	if err != nil {
		return nil, err
	}
	return &Sprint{target: sprint}, nil
}

// GetWait gets the current wait of this session.. can't call this Wait() because Object in Java already has a wait() method
func (s *Session) GetWait() *Wait {
	if s.target.Wait() != nil {
		return &Wait{target: s.target.Wait()}
	}
	return nil
}

// ToJSON serializes this session as JSON
func (s *Session) ToJSON() (string, error) {
	data, err := json.Marshal(s.target)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type Hint struct {
	target flows.Hint
}

func (h *Hint) Type() string {
	return string(h.target.Type())
}

type Wait struct {
	target flows.Wait
}

func (w *Wait) Type() string {
	return string(w.target.Type())
}

func (w *Wait) Hint() *Hint {
	asMsgWait, isMsgWait := w.target.(*waits.MsgWait)
	if isMsgWait && asMsgWait.Hint() != nil {
		return &Hint{target: asMsgWait.Hint()}
	}
	return nil
}

type Engine struct {
	target flows.Engine
}

func NewEngine(httpUserAgent string) *Engine {
	return &Engine{
		target: engine.NewBuilder().WithDefaultUserAgent(httpUserAgent).Build(),
	}
}

// NewSession creates a new session
func (e *Engine) NewSession(sa *SessionAssets) *Session {
	return &Session{target: e.target.NewSession(sa.target)}
}

// ReadSession reads an existing session from JSON
func (e *Engine) ReadSession(a *SessionAssets, data string) (*Session, error) {
	s, err := e.target.ReadSession(a.target, []byte(data), assets.IgnoreMissing)
	if err != nil {
		return nil, err
	}
	return &Session{target: s}, nil
}
