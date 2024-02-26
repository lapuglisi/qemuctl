package qemuctl_actions

import "fmt"

type GenericAction interface {
	Run(arguments []string) error
}

type DummyAction struct {
}

var actionsMap map[string]GenericAction

func init() {
	actionsMap = make(map[string]GenericAction, 0)

	actionsMap["attach"] = &AttachAction{}
	actionsMap["completion"] = &CompletionAction{}
	actionsMap["create"] = &CreateAction{}
	actionsMap["destroy"] = &DestroyAction{}
	actionsMap["disable"] = &DisableAction{}
	actionsMap["enable"] = &EnableAction{}
	actionsMap["help"] = &HelpAction{}
	actionsMap["info"] = &InfoAction{}
	actionsMap["kill"] = &KillAction{}
	actionsMap["list"] = &ListAction{}
	actionsMap["service"] = &ServiceAction{}
	actionsMap["start"] = &StartAction{}
	actionsMap["status"] = &StatusAction{}
	actionsMap["stop"] = &StopAction{}
}

func GetActionInterface(action string) (out GenericAction) {
	final := actionsMap[action]
	if final == nil {
		return &DummyAction{}
	}

	return final
}

func GetAllActionStrings() (list []string) {
	list = make([]string, 0)

	for action := range actionsMap {
		list = append(list, action)
	}

	return list
}

func (dummy *DummyAction) Run(arguments []string) (err error) {
	return fmt.Errorf("dummy action called")
}
