package tg

func notice(msg string) {
	go api.SendMessage(msg)
}