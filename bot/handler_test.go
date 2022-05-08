package bot

import (
	"strconv"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var config = Config{
	Chats: []Chat{
		{ID: 1, Aliases: []string{"First", "All", "SingleDigit"},
			ChildChats: []Chat{
				{ID: 10, Aliases: []string{"Tenth", "All", "DoubleDigit"},
					ChildChats: []Chat{{ID: 100, Aliases: []string{"Hundreadth", "All", "TripleDigit"}}}},
				{ID: 11, Aliases: []string{"Eleventh", "All", "DoubleDigit"}},
			},
		},
		{ID: 2, Aliases: []string{"Second", "All", "SingleDigit"}},
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
	for _, testCase := range []struct {
		name         string
		fromChatID   int64
		messageID    int
		messageText  string
		wantForwards []tgbotapi.ForwardConfig
	}{
		{name: "Resend between siblings: from First to Second",
			fromChatID:  1,
			messageID:   152,
			messageText: "My message *second",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 152, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 2}},
			},
		},
		{name: "No tags doesn't resend the message",
			fromChatID:   1,
			messageID:    651,
			messageText:  "My message",
			wantForwards: nil,
		},
		{name: "Infix tag",
			fromChatID:  1,
			messageID:   781,
			messageText: "My *second message",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 781, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 2}},
			},
		},
		{name: "Resend to siblings alias: from First to SingleDigit",
			fromChatID:  1,
			messageID:   186,
			messageText: "My message *SingleDigit",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 186, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 1}},
				{MessageID: 186, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 2}},
			},
		},
		{name: "Resend to child chats: from First to DoubleDigit",
			fromChatID:  1,
			messageID:   785,
			messageText: "My message *DoubleDigit",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 785, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 10}},
				{MessageID: 785, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 11}},
			},
		},
		{name: "Resend to newphew chats from Second to DoubleDigit",
			fromChatID:  2,
			messageID:   734,
			messageText: "My message *DoubleDigit",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 734, FromChatID: 2, BaseChat: tgbotapi.BaseChat{ChatID: 10}},
				{MessageID: 734, FromChatID: 2, BaseChat: tgbotapi.BaseChat{ChatID: 11}},
			},
		},
		{name: "Resend to the grandchild chat: from First to TripleDigit",
			fromChatID:  1,
			messageID:   679,
			messageText: "My message *TripleDigit",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 679, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 100}},
			},
		},
		{name: "Resend to the grandnewphew chat: from Second to TripleDigit",
			fromChatID:  2,
			messageID:   813,
			messageText: "My message *TripleDigit",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 813, FromChatID: 2, BaseChat: tgbotapi.BaseChat{ChatID: 100}},
			},
		},
		{name: "Resend to the parent chat: from Tenth to First",
			fromChatID:  10,
			messageID:   519,
			messageText: "My message *first",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 519, FromChatID: 10, BaseChat: tgbotapi.BaseChat{ChatID: 1}},
			},
		},
		{name: "Resend to the aunt chat: from Tenth to Second",
			fromChatID:  10,
			messageID:   90,
			messageText: "My message *second",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 90, FromChatID: 10, BaseChat: tgbotapi.BaseChat{ChatID: 2}},
			},
		},
		{name: "Resend to the grandparent chat: from Hundreadth to First",
			fromChatID:  100,
			messageID:   515,
			messageText: "My message *first",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 515, FromChatID: 100, BaseChat: tgbotapi.BaseChat{ChatID: 1}},
			},
		},
		{name: "Resend to the grandaunt chat: from Hundreadth to Second",
			fromChatID:  100,
			messageID:   844,
			messageText: "My message *second",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 844, FromChatID: 100, BaseChat: tgbotapi.BaseChat{ChatID: 2}},
			},
		},
		{name: "Resend to the grandparent alias: from Hundreadths to SingleDigit",
			fromChatID:  100,
			messageID:   960,
			messageText: "My message *SingleDigit",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 960, FromChatID: 100, BaseChat: tgbotapi.BaseChat{ChatID: 1}},
				{MessageID: 960, FromChatID: 100, BaseChat: tgbotapi.BaseChat{ChatID: 2}},
			},
		},
		{name: "Resend to the parent alias: from Hundreadths to DoubleDigit",
			fromChatID:  100,
			messageID:   177,
			messageText: "My message *DoubleDigit",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 177, FromChatID: 100, BaseChat: tgbotapi.BaseChat{ChatID: 10}},
				{MessageID: 177, FromChatID: 100, BaseChat: tgbotapi.BaseChat{ChatID: 11}},
			},
		},
		{name: "Resend to multiple aliases: from First to DoulbeDigit+TripleDigit",
			fromChatID:  1,
			messageID:   880,
			messageText: "My message *DoubleDigit *TripleDigit",
			wantForwards: []tgbotapi.ForwardConfig{
				{MessageID: 880, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 10}},
				{MessageID: 880, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 11}},
				{MessageID: 880, FromChatID: 1, BaseChat: tgbotapi.BaseChat{ChatID: 100}},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			bot := &fakeBot{}
			handler := NewHandler(config, bot)

			handler.HandleUpdate(tgbotapi.Update{
				Message: &tgbotapi.Message{
					Chat:      &tgbotapi.Chat{ID: testCase.fromChatID},
					From:      &tgbotapi.User{},
					MessageID: testCase.messageID,
					Text:      testCase.messageText,
				},
			})

			var forwards []tgbotapi.ForwardConfig
			for _, chattable := range bot.sentMessages {
				if fc, ok := chattable.(tgbotapi.ForwardConfig); ok {
					forwards = append(forwards, fc)
				}
			}
			diff := cmp.Diff(testCase.wantForwards, forwards, cmpopts.SortSlices(func(one, another tgbotapi.ForwardConfig) bool {
				if one.FromChatID != another.FromChatID {
					return one.FromChatID < another.FromChatID
				}
				return one.MessageID < another.MessageID
			}))
			if diff != "" {
				t.Fatalf("The bot sent wrong forwards, cmp.Diff(want, got): %s", diff)
			}
		})
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
				tgbotapi.NewInlineQueryResultArticle("*DoubleDigit", "*DoubleDigit", "*DoubleDigit"),
				tgbotapi.NewInlineQueryResultArticle("*Eleventh", "*Eleventh", "*Eleventh"),
				tgbotapi.NewInlineQueryResultArticle("*First", "*First", "*First"),
				tgbotapi.NewInlineQueryResultArticle("*Hundreadth", "*Hundreadth", "*Hundreadth"),
				tgbotapi.NewInlineQueryResultArticle("*Second", "*Second", "*Second"),
				tgbotapi.NewInlineQueryResultArticle("*SingleDigit", "*SingleDigit", "*SingleDigit"),
				tgbotapi.NewInlineQueryResultArticle("*Tenth", "*Tenth", "*Tenth"),
				tgbotapi.NewInlineQueryResultArticle("*TripleDigit", "*TripleDigit", "*TripleDigit"),
			}},
		{name: "A star suggests all tags",
			query: "*",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("*All", "*All", "*All"),
				tgbotapi.NewInlineQueryResultArticle("*DoubleDigit", "*DoubleDigit", "*DoubleDigit"),
				tgbotapi.NewInlineQueryResultArticle("*Eleventh", "*Eleventh", "*Eleventh"),
				tgbotapi.NewInlineQueryResultArticle("*First", "*First", "*First"),
				tgbotapi.NewInlineQueryResultArticle("*Hundreadth", "*Hundreadth", "*Hundreadth"),
				tgbotapi.NewInlineQueryResultArticle("*Second", "*Second", "*Second"),
				tgbotapi.NewInlineQueryResultArticle("*SingleDigit", "*SingleDigit", "*SingleDigit"),
				tgbotapi.NewInlineQueryResultArticle("*Tenth", "*Tenth", "*Tenth"),
				tgbotapi.NewInlineQueryResultArticle("*TripleDigit", "*TripleDigit", "*TripleDigit"),
			}},
		{name: "A star and a subword filters suggestions",
			query: "*Se",
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
				tgbotapi.NewInlineQueryResultArticle("Some message *DoubleDigit", "*DoubleDigit", "Some message *DoubleDigit"),
				tgbotapi.NewInlineQueryResultArticle("Some message *Eleventh", "*Eleventh", "Some message *Eleventh"),
				tgbotapi.NewInlineQueryResultArticle("Some message *First", "*First", "Some message *First"),
				tgbotapi.NewInlineQueryResultArticle("Some message *Hundreadth", "*Hundreadth", "Some message *Hundreadth"),
				tgbotapi.NewInlineQueryResultArticle("Some message *Second", "*Second", "Some message *Second"),
				tgbotapi.NewInlineQueryResultArticle("Some message *SingleDigit", "*SingleDigit", "Some message *SingleDigit"),
				tgbotapi.NewInlineQueryResultArticle("Some message *Tenth", "*Tenth", "Some message *Tenth"),
				tgbotapi.NewInlineQueryResultArticle("Some message *TripleDigit", "*TripleDigit", "Some message *TripleDigit"),
			}},
		{name: "A star after some text immediately without any space doesn't suggest anything to not spam",
			query:       "Some message*",
			wantResults: nil},
		{name: "A star after a tag and a space suggests all tags",
			query: "*first *",
			wantResults: []interface{}{
				tgbotapi.NewInlineQueryResultArticle("*first *All", "*All", "*first *All"),
				tgbotapi.NewInlineQueryResultArticle("*first *DoubleDigit", "*DoubleDigit", "*first *DoubleDigit"),
				tgbotapi.NewInlineQueryResultArticle("*first *Eleventh", "*Eleventh", "*first *Eleventh"),
				tgbotapi.NewInlineQueryResultArticle("*first *First", "*First", "*first *First"),
				tgbotapi.NewInlineQueryResultArticle("*first *Hundreadth", "*Hundreadth", "*first *Hundreadth"),
				tgbotapi.NewInlineQueryResultArticle("*first *Second", "*Second", "*first *Second"),
				tgbotapi.NewInlineQueryResultArticle("*first *SingleDigit", "*SingleDigit", "*first *SingleDigit"),
				tgbotapi.NewInlineQueryResultArticle("*first *Tenth", "*Tenth", "*first *Tenth"),
				tgbotapi.NewInlineQueryResultArticle("*first *TripleDigit", "*TripleDigit", "*first *TripleDigit"),
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
