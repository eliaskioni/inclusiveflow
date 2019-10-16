package actions

import (
	"encoding/json"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/pkg/errors"
)

func init() {
	registerType(TypeCallClassifier, func() flows.Action { return &CallClassifierAction{} })
}

var classificationCategories = []string{CategorySuccess, CategorySkipped, CategoryFailure}

// TypeCallClassifier is the type for the call classifier action
const TypeCallClassifier string = "call_classifier"

// CallClassifierAction can be used to classify the intent and entities from a given input using an NLU classifier. It always
// saves a result indicating whether the classification was successful, skipped or failed, and what the extracted intents
// and entities were.
//
//   {
//     "uuid": "8eebd020-1af5-431c-b943-aa670fc74da9",
//     "type": "call_classifier",
//     "classifier": {
//       "uuid": "1c06c884-39dd-4ce4-ad9f-9a01cbe6c000",
//       "name": "Booking"
//     },
//     "input": "@input.text",
//     "result_name": "Intent"
//   }
//
// @action call_classifier
type CallClassifierAction struct {
	baseAction
	onlineAction

	Classifier *assets.ClassifierReference `json:"classifier" validate:"required"`
	Input      string                      `json:"input" validate:"required" engine:"evaluated"`
	ResultName string                      `json:"result_name" validate:"required"`
}

// NewCallClassifier creates a new call classifier action
func NewCallClassifier(uuid flows.ActionUUID, classifier *assets.ClassifierReference, input string, resultName string) *CallClassifierAction {
	return &CallClassifierAction{
		baseAction: newBaseAction(TypeCallClassifier, uuid),
		Classifier: classifier,
		Input:      input,
		ResultName: resultName,
	}
}

// Execute runs this action
func (a *CallClassifierAction) Execute(run flows.FlowRun, step flows.Step, logModifier flows.ModifierCallback, logEvent flows.EventCallback) error {
	classifiers := run.Session().Assets().Classifiers()
	classifier := classifiers.Get(a.Classifier.UUID)

	// substitute any variables in our input
	input, err := run.EvaluateTemplate(a.Input)
	if err != nil {
		logEvent(events.NewError(err))
	}

	classification, skipped, err := a.classify(run, step, input, classifier, logEvent)
	if err != nil {
		logEvent(events.NewError(err))

		if skipped {
			a.saveSkipped(run, step, input, logEvent)
		} else {
			a.saveFailure(run, step, input, logEvent)
		}
	} else {
		a.saveSuccess(run, step, input, classification, logEvent)
	}

	return nil
}

func (a *CallClassifierAction) classify(run flows.FlowRun, step flows.Step, input string, classifier *flows.Classifier, logEvent flows.EventCallback) (*flows.Classification, bool, error) {
	if input == "" {
		return nil, true, errors.New("can't classify empty input, skipping classification")
	}
	if classifier == nil {
		return nil, false, errors.Errorf("missing %s", a.Classifier.String())
	}

	svc, err := run.Session().Engine().Services().Classification(run.Session(), classifier)
	if err != nil {
		return nil, false, err
	}

	httpLogger := &flows.HTTPLogger{}

	classification, err := svc.Classify(run.Session(), input, httpLogger.Log)

	if len(httpLogger.Logs) > 0 {
		logEvent(events.NewClassifierCalled(classifier.Reference(), httpLogger.Logs))
	}

	return classification, false, err
}

func (a *CallClassifierAction) saveSuccess(run flows.FlowRun, step flows.Step, input string, classification *flows.Classification, logEvent flows.EventCallback) {
	// result value is name of top ranked intent if there is one
	value := ""
	if len(classification.Intents) > 0 {
		value = classification.Intents[0].Name
	}
	extra, _ := json.Marshal(classification)

	a.saveResult(run, step, a.ResultName, value, CategorySuccess, "", input, extra, logEvent)
}

func (a *CallClassifierAction) saveSkipped(run flows.FlowRun, step flows.Step, input string, logEvent flows.EventCallback) {
	a.saveResult(run, step, a.ResultName, "0", CategorySkipped, "", input, nil, logEvent)
}

func (a *CallClassifierAction) saveFailure(run flows.FlowRun, step flows.Step, input string, logEvent flows.EventCallback) {
	a.saveResult(run, step, a.ResultName, "0", CategoryFailure, "", input, nil, logEvent)
}

// Results enumerates any results generated by this flow object
func (a *CallClassifierAction) Results(node flows.Node, include func(*flows.ResultInfo)) {
	if a.ResultName != "" {
		include(flows.NewResultInfo(a.ResultName, classificationCategories, node))
	}
}
