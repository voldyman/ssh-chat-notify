package notifyi

type Comms interface {
	PrivateMessage(toUsername, message string) error
	PublicMessage(message string) error
}

type Bot struct {
	name  string
	comms Comms
}

func New(name string, comms Comms) *Bot {
	return &Bot{
		name:  name,
		comms: comms,
	}
}

func (b *Bot) PublicMessage(username, message string) error {
	return nil
}

func (b *Bot) PrivateMessage(username, message string) error {
	cmd := parsePrivateMessage(message)
	if cmd == nil {
		//b.comms.PrivateMessage(username, "i am still a WIP, talk to voldyman if you want to know more")
		return b.sendHelp(username)
	}
	return nil
}

func (b *Bot) ActionMessage(username, action string) error {
	return nil
}

func (b *Bot) UserJoinedMessage(username string) error {
	return nil
}

func (b *Bot) UserLeftMessage(username string) error {
	return nil
}

func (b *Bot) UsernameChangeMessage(from, to string) error {
	return nil
}

func (b *Bot) sendHelp(username string) error {
	help := helpCmd{myusername: b.name, sendTo: username}
	return help.Execute(b.comms)
}
