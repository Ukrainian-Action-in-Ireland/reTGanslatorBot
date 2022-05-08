package bot

import (
	"strconv"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var config = Config{
	Chats: []Chat{
		{ID: 1, Aliases: []string{"First", "All"}},
		{ID: 2, Aliases: []string{"Second", "All"}},
	},
	HelpContacts: []string{"@Kyslytsya", "@Karas", "@Valera", "@Arestovich"},
}

type fakeBot struct {
	sentMessages []tgbotapi.Chattable
	inlineConfig tgbotapi.InlineConfig
}

func (fb *fakeBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	fb.sentMessages = append(fb.sentMessages, c)
	return tgbotapi.Message{}, nil
}

func (fb *fakeBot) AnswerInlineQuery(config tgbotapi.InlineConfig) (tgbotapi.APIResponse, error) {
	return tgbotapi.APIResponse{}, nil
}

func TestTagResendsMessage(t *testing.T) {
	bot := &fakeBot{}
	handler := NewHandler(config, bot)

	handler.HandleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat:      &tgbotapi.Chat{ID: 1},
			From:      &tgbotapi.User{},
			MessageID: 152,
			Text:      "Message text *second",
		},
	})

	var forwards []tgbotapi.ForwardConfig
	for _, chattable := range bot.sentMessages {
		if fc, ok := chattable.(tgbotapi.ForwardConfig); ok {
			forwards = append(forwards, fc)
		}
	}
	if len(forwards) != 1 {
		t.Errorf("Expected 1 forwarded message, got %v forward messages. Messages:", len(forwards))
		for _, chattable := range bot.sentMessages {
			t.Errorf("Chattable: %v, %T", chattable, chattable)
		}
		t.FailNow()
	}
	forward := forwards[0]
	if forward.FromChatID != 1 {
		t.Errorf("Expected FromChatID == 1, got %v", forward.FromChatID)
	}
	if forward.ChatID != 2 {
		t.Errorf("Expected ChatID == 2, got %v", forward.ChatID)
	}
	if forward.MessageID != 152 {
		t.Errorf("Expected MessageID == 152, got %v", forward.MessageID)
	}
}

func TestNoTagsDoNotResendMessage(t *testing.T) {
	bot := &fakeBot{}
	handler := NewHandler(config, bot)

	handler.HandleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat:      &tgbotapi.Chat{ID: 1},
			From:      &tgbotapi.User{},
			MessageID: 152,
			Text:      "Message text second",
		},
	})

	if bot.sentMessages != nil {
		t.Error("Expected no messages, got:")
		for _, chattable := range bot.sentMessages {
			t.Errorf("Chattable: %v, %T", chattable, chattable)
		}
		t.FailNow()
	}
}

func TestMultipleAliasSendsToMultipleChats(t *testing.T) {
	bot := &fakeBot{}
	handler := NewHandler(config, bot)

	handler.HandleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat:      &tgbotapi.Chat{ID: 2},
			From:      &tgbotapi.User{},
			MessageID: 152,
			Text:      "Message text *all",
		},
	})

	var forwards []tgbotapi.ForwardConfig
	for _, chattable := range bot.sentMessages {
		if fc, ok := chattable.(tgbotapi.ForwardConfig); ok {
			forwards = append(forwards, fc)
		}
	}
	if len(forwards) != 2 {
		t.Errorf("Expected 2 forwarded message, got %v forward messages. Messages:", len(forwards))
		for _, chattable := range bot.sentMessages {
			t.Errorf("Chattable: %v, %T", chattable, chattable)
		}
		t.FailNow()
	}
	sentToMap := map[int64]bool{}
	for _, forward := range forwards {
		if forward.FromChatID != 2 {
			t.Errorf("Expected all messages to be forwarded from the chat with ID=2, got: %v. Message: %v %T", forward.FromChatID, forward, forward)
		}
		sentToMap[forward.ChatID] = true
	}
	var sentTo []int64
	var sentToStrs []string
	for id := range sentToMap {
		sentTo = append(sentTo, id)
		sentToStrs = append(sentToStrs, strconv.FormatInt(id, 10))
	}
	if len(sentTo) != 2 || !sentToMap[1] || !sentToMap[2] {
		t.Errorf("Expected forwarded messages in chat with IDs 1, 2; got chats with IDs: %v", strings.Join(sentToStrs, ", "))
	}
}

func TestReplyResendsBothMessages(t *testing.T) {
	bot := &fakeBot{}
	handler := NewHandler(config, bot)

	origMsg := &tgbotapi.Message{
		Chat:      &tgbotapi.Chat{ID: 1},
		From:      &tgbotapi.User{},
		MessageID: 152,
		Text:      "Message text",
	}
	handler.HandleUpdate(tgbotapi.Update{Message: origMsg})
	handler.HandleUpdate(tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat:           &tgbotapi.Chat{ID: 1},
			From:           &tgbotapi.User{},
			MessageID:      612,
			ReplyToMessage: origMsg,
			Text:           "P.S. pls, resend *second",
		},
	})

	var forwards []tgbotapi.ForwardConfig
	for _, chattable := range bot.sentMessages {
		if fc, ok := chattable.(tgbotapi.ForwardConfig); ok {
			forwards = append(forwards, fc)
		}
	}
	if len(forwards) != 2 {
		t.Errorf("Expected 2 forwarded message, got %v forward messages. Messages:", len(forwards))
		for _, chattable := range bot.sentMessages {
			t.Errorf("Chattable: %v, %T", chattable, chattable)
		}
		t.FailNow()
	}
	sentMap := map[int]bool{}
	for _, forward := range forwards {
		if forward.FromChatID != 1 {
			t.Errorf("Expected all messages to be forwarded from the chat with ID=1, got: %v. Message: %v %T", forward.FromChatID, forward, forward)
		}
		sentMap[forward.MessageID] = true
	}
	var sent []int
	var sentStrs []string
	for id := range sentMap {
		sent = append(sent, id)
		sentStrs = append(sentStrs, strconv.Itoa(id))
	}
	if len(sent) != 2 {
		t.Errorf("Expected 2 messages with IDs 152, 612; got with IDs %v", strings.Join(sentStrs, ", "))
	}
}
