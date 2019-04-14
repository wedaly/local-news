package i18n

// MsgId identifies a user-facing string that can be translated.
type MsgId string

func InitTranslations(domain string) {
	// TODO
}

func Gettext(msgId MsgId) string {
	// TODO
	return string(msgId)
}

func NGettext(singularMsgId MsgId, pluralMsgId MsgId, count int) string {
	// TODO
	if count == 1 {
		return string(singularMsgId)
	} else {
		return string(pluralMsgId)
	}
}
