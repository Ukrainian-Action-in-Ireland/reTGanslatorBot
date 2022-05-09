package bot

import (
	"strconv"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/go-cmp/cmp"
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
	fb.inlineConfig = config
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

func TestInlineQueries(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		query       string
		wantResults []interface{}
	}{
		{name: "Empty query suggests all tags",
			query: "",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("*All", "*All", "*All"),
				tgbotapi.NewInlineQueryResultArticle("*First", "*First", "*First"),
				tgbotapi.NewInlineQueryResultArticle("*Second", "*Second", "*Second"),
			}},
		{name: "A star suggests all tags",
			query: "*",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("*All", "*All", "*All"),
				tgbotapi.NewInlineQueryResultArticle("*First", "*First", "*First"),
				tgbotapi.NewInlineQueryResultArticle("*Second", "*Second", "*Second"),
			}},
		{name: "A star and a subword filters suggestions",
			query: "*S",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("*Second", "*Second", "*Second"),
			}},
		{name: "A subword without matches gives no suggestions",
			query:       "*thi",
			wantResults: nil},
		{name: "A subword without star filters suggestions",
			query: "Fir",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("*First", "*First", "*First"),
			}},
		{name: "A subword with a different letter case still matches",
			query: "sec",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("*Second", "*Second", "*Second"),
			}},
		{name: "A star after some text and some space suggests all tags",
			query: "Some message *",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("Some message *All", "*All", "Some message *All"),
				tgbotapi.NewInlineQueryResultArticle("Some message *First", "*First", "Some message *First"),
				tgbotapi.NewInlineQueryResultArticle("Some message *Second", "*Second", "Some message *Second"),
			}},
		{name: "A star after some text immediately without any space doesn't suggest anything to not spam",
			query:       "Some message*",
			wantResults: nil},
		{name: "A star after a tag and a space suggests all tags",
			query: "*first *",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("*first *All", "*All", "*first *All"),
				tgbotapi.NewInlineQueryResultArticle("*first *First", "*First", "*first *First"),
				tgbotapi.NewInlineQueryResultArticle("*first *Second", "*Second", "*first *Second"),
			}},
		{name: "A star and a subword after a tag and a space filters the tags",
			query: "*first *a",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("*first *All", "*All", "*first *All"),
			}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			bot := &fakeBot{}
			handler := NewHandler(config, bot)

			handler.HandleUpdate(tgbotapi.Update{InlineQuery: &tgbotapi.InlineQuery{
				Query: testCase.query,
			}})

			if diff := cmp.Diff(testCase.wantResults, bot.inlineConfig.Results); diff != "" {
				t.Fatalf("Got incorrect inline query results, cmp.Diff(want, got):\n %s", diff)
			}
		})
	}
}
