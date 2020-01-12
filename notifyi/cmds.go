package notifyi

import "fmt"

const helpCmdName = "help"
const registerCmdName = "register"
const verifyCmdName = "verify"
const addWatchCmdName = "add-watch"
const stopWatchCmdName = "stop-watch"

type executableCmd interface {
	Execute(responder Comms) error
}

type helpCmd struct {
	myusername string
	sendTo     string
}

func (h *helpCmd) Execute(comms Comms) error {
	helpMessage := []string{
		fmt.Sprintf("/msg %s %s", h.myusername, helpCmdName),
		fmt.Sprintf("/msg %s %s <email>", h.myusername, registerCmdName),
		fmt.Sprintf("/msg %s %s <email> <verification code>", h.myusername, verifyCmdName),
		fmt.Sprintf("/msg %s %s <token>", h.myusername, addWatchCmdName),
		fmt.Sprintf("/msg %s %s <token>", h.myusername, stopWatchCmdName),
	}
	for _, msg := range helpMessage {
		err := comms.PrivateMessage(h.sendTo, msg)
		if err != nil {
			return fmt.Errorf("unable to send help to user %s: %w", h.sendTo, err)
		}
	}
	return nil
}

type registerCmd struct{}
type verifyCmd struct{}
type addWatchCmd struct{}
type stopWatchCmd struct{}
