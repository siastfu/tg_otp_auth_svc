package telegram

import (
	"fmt"
	"log"
	"tg_otp_auth_svc/internal/infra/usecase"
	"time"

	"gopkg.in/telebot.v3"
	"tg_otp_auth_svc/internal/domain"
)

// TelegramHandler - структура для работы с Telegram-ботом
type TelegramHandler struct {
	Bot    *telebot.Bot
	AuthUC domain.AuthUseCase
}

// NewTelegramHandler - создание нового Telegram-хендлера
func NewTelegramHandler(authUC *usecase.AuthUseCaseImpl, botToken string) (*TelegramHandler, error) {
	pref := telebot.Settings{
		Token:  botToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании бота: %w", err)
	}

	handler := &TelegramHandler{
		Bot:    bot,
		AuthUC: authUC,
	}

	log.Println("✅ Бот успешно создан!")

	// Регистрируем обработчики
	handler.registerHandlers()

	return handler, nil
}

// registerHandlers - регистрация команд бота
func (h *TelegramHandler) registerHandlers() {
	h.Bot.Handle(&telebot.Command{Text: "start"}, h.handleStart)
	h.Bot.Handle(&telebot.Callback{Unique: "auth"}, h.handleAuth)
	h.Bot.Handle(&telebot.Callback{Unique: "lang_ru"}, func(c telebot.Context) error {
		return h.updateUserLanguage(c, "ru")
	})
	h.Bot.Handle(&telebot.Callback{Unique: "lang_en"}, func(c telebot.Context) error {
		return h.updateUserLanguage(c, "en")
	})
}

// handleStart - обработчик команды /start
func (h *TelegramHandler) handleStart(c telebot.Context) error {
	tgID := c.Sender().ID

	// Проверяем, есть ли пользователь в БД
	user, err := h.AuthUC.GetUserByTgID(tgID)
	if err != nil {
		log.Println("Ошибка при получении пользователя:", err)
		return c.Send("Произошла ошибка. Попробуйте еще раз.")
	}

	if user == nil {
		// Если пользователя нет — создаем запись
		newUser := &domain.User{
			ChatID:   tgID,
			Username: c.Sender().Username,
		}

		err = h.AuthUC.CreateUser(newUser)
		if err != nil {
			log.Println("Ошибка при создании пользователя:", err)
			return c.Send("Ошибка при создании пользователя.")
		}

		// Предлагаем выбрать язык
		return c.Send("Выберите язык:", &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{{Text: "🇷🇺 Русский", Unique: "lang_ru"}},
				{{Text: "🇬🇧 English", Unique: "lang_en"}},
			},
		})
	}

	// Если пользователь уже есть — показываем основное меню
	return h.showMainMenu(c)
}

// handleAuth - обработчик кнопки "Авторизоваться"
func (h *TelegramHandler) handleAuth(c telebot.Context) error {
	tgID := c.Sender().ID

	// Проверяем, есть ли активная авторизационная ссылка
	authLink, err := h.AuthUC.StartAuthorization(tgID)
	if err != nil {
		log.Println("Ошибка авторизации:", err)
		return c.Send("Ошибка при авторизации. Попробуйте снова.")
	}

	// Отправляем ссылку пользователю
	return c.Send("Перейдите по ссылке для авторизации:", &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{{Text: "🔑 Войти", URL: authLink}},
		},
	})
}

// updateUserLanguage - обновляет язык пользователя в БД
func (h *TelegramHandler) updateUserLanguage(c telebot.Context, lang string) error {
	tgID := c.Sender().ID

	err := h.AuthUC.UpdateUserLanguage(tgID, lang)
	if err != nil {
		log.Println("Ошибка обновления языка:", err)
		return c.Send("Ошибка при смене языка.")
	}

	return c.Send("Язык успешно изменен! Теперь все сообщения будут отображаться на вашем языке.", h.showMainMenu(c))
}

// showMainMenu - отправляет основное меню
func (h *TelegramHandler) showMainMenu(c telebot.Context) error {
	return c.Send("Выберите действие:", &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{{Text: "🔑 Авторизоваться", Unique: "auth"}},
			{{Text: "🌍 Язык", Unique: "change_lang"}},
		},
	})
}

// Run - запуск Telegram-бота
func (h *TelegramHandler) Run() {
	log.Println("✅ Telegram-бот запущен!")
	h.Bot.Start()
}
