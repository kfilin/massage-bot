package handlers

import (
	"context"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kfilin/massage-bot/internal/domain"
	"github.com/kfilin/massage-bot/internal/logging"
	"gopkg.in/telebot.v3"
)

// File/media/voice handling on BookingHandler. Handles document uploads
// from patients, voice transcription with draft approval, and the
// admin-reply flow.
func (h *BookingHandler) HandleUploadCommand(c telebot.Context) error {
	return c.Send(`📤 *Загрузка медицинских документов*

Вы можете отправить мне свои результаты обследований (МРТ, КТ, рентген, анализы) в форматах **PDF, JPG, PNG** или **DICOM (.dcm)**.

*Инструкция:*
1. Просто прикрепите файл или фото к сообщению и отправьте его мне.
2. Я автоматически сохраню его в вашу медицинскую карту.
3. Доктор увидит ваши документы при следующем посещении.

⚠️ *Максимальный размер файла: 20 МБ (Ограничение Telegram)*`, telebot.ParseMode(telebot.ModeMarkdown))
}

func (h *BookingHandler) HandleFileMessage(c telebot.Context) error {
	userID := c.Sender().ID
	telegramID := strconv.FormatInt(userID, 10)

	var fileID string
	var fileName string
	var fileSize int

	msg := c.Message()
	if doc := msg.Document; doc != nil {
		fileID = doc.FileID
		fileName = doc.FileName
		fileSize = int(doc.FileSize)
	} else if photo := msg.Photo; photo != nil {
		fileID = photo.FileID
		fileName = fmt.Sprintf("photo_%d.jpg", time.Now().Unix())
		fileSize = int(photo.FileSize)
	} else if vid := msg.Video; vid != nil {
		fileID = vid.FileID
		fileName = vid.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("video_%d.mp4", time.Now().Unix())
		}
		fileSize = int(vid.FileSize)
	} else if anim := msg.Animation; anim != nil {
		fileID = anim.FileID
		fileName = anim.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("animation_%d.mp4", time.Now().Unix())
		}
		fileSize = int(anim.FileSize)
	} else if voice := msg.Voice; voice != nil {
		fileID = voice.FileID
		fileName = fmt.Sprintf("voice_%d.ogg", time.Now().Unix())
		fileSize = int(voice.FileSize)
	} else {
		return nil // Not a recognized media type
	}

	// 20MB limit for Public Telegram API
	if fileSize > 20*1024*1024 {
		return c.Send("❌ Файл слишком большой. Максимальный размер: 20 МБ (Ограничение Telegram).")
	}

	// Check if patient exists
	patient, err := h.repository.GetPatient(telegramID)
	if err != nil {
		return c.Send("❌ Сначала запишитесь на прием через /start, чтобы я мог создать вашу карту и сохранить документ.")
	}

	statusMsg, err := c.Bot().Send(c.Recipient(), "⏳ Загружаю и сохраняю ваш файл...")
	if err != nil {
		logging.Errorf(": Failed to send status message: %v", err)
	}

	// Get file from Telegram servers
	fileReader, err := c.Bot().File(&telebot.File{FileID: fileID})
	if err != nil {
		logging.Errorf(": Failed to download file from Telegram: %v", err)
		if statusMsg != nil {
			if err := c.Bot().Delete(statusMsg); err != nil {
				logging.Warnf("Failed to delete status message: %v", err)
			}
		}
		return c.Send("❌ Ошибка при загрузке файла. Возможно, он слишком большой для Telegram-бота (лимит 50МБ).\n\nПопробуйте отправить файл меньшего размера или ссылкой.")
	}
	defer fileReader.Close()

	// 1. Check if this is an Admin replying to a patient
	session := h.sessionStorage.Get(userID)
	if replyingToID, ok := session[SessionKeyAdminReplyingTo].(string); ok && replyingToID != "" {
		logging.Infof("[Reply] Admin %d is replying to patient %s via file/voice", userID, replyingToID)

		patientID, _ := strconv.ParseInt(replyingToID, 10, 64)
		patientUser := &telebot.User{ID: patientID}

		// Forward the file/voice itself
		_, err := c.Bot().Copy(patientUser, c.Message())
		if err != nil {
			logging.Errorf("Failed to forward voice/file to patient %s: %v", replyingToID, err)
			return c.Send("❌ Не удалось отправить файл пациенту.")
		}

		// If it's a voice message, transcribe it and send text too
		var transcript string
		if voice := msg.Voice; voice != nil {
			statusMsg, _ := c.Bot().Send(c.Sender(), "📝 Расшифровываю ваш ответ...")

			// Need a new fileReader as the previous one was closed by defer
			fileReaderForTranscription, _ := c.Bot().File(&telebot.File{FileID: voice.FileID})
			defer fileReaderForTranscription.Close() // Ensure this is closed too

			// Use a generic name for admin replies to avoid confusion
			transcript, err = h.transcriptionService.Transcribe(context.Background(), fileReaderForTranscription, "admin_reply.ogg")

			if statusMsg != nil {
				if err := c.Bot().Delete(statusMsg); err != nil {
					logging.Warnf("Failed to delete status message: %v", err)
				}
			}

			if err == nil && transcript != "" {
				// Send transcription to patient
				if _, err := c.Bot().Send(patientUser, fmt.Sprintf("📝 <b>Текст сообщения:</b>\n%s", transcript), telebot.ModeHTML); err != nil {
					logging.Warnf("Failed to send transcript to patient: %v", err)
				}

				// Log to Patient's Notes (Dialogue View)
				patient, err := h.repository.GetPatient(replyingToID)
				if err == nil {
					// Add date header if this is the first message of the day in notes
					today := time.Now().In(domain.ApptTimeZone).Format("02.01.2006")
					dateHeader := fmt.Sprintf("\n\n📅 %s", today)
					if !strings.Contains(patient.TherapistNotes, dateHeader) {
						patient.TherapistNotes += dateHeader
					}

					notePrefix := fmt.Sprintf("\n\n[🗣 Вера %s]: ", time.Now().In(domain.ApptTimeZone).Format("15:04"))
					patient.TherapistNotes += notePrefix + transcript
					if err := h.repository.SavePatient(patient); err != nil {
						logging.Errorf("Failed to save admin reply to patient record: %v", err)
					}
				}
			}
		}

		// Clear session
		h.sessionStorage.Set(userID, SessionKeyAdminReplyingTo, nil)
		return c.Send(fmt.Sprintf("✅ Сообщение отправлено пациенту (ID: %s)", replyingToID))
	}

	// 2. Standard Flow: Patient Uploading File
	// Determine category based on extension/type
	ext := strings.ToLower(filepath.Ext(fileName))

	// Determine file type for DB
	fileType := "document"
	if msg.Voice != nil || msg.Audio != nil {
		fileType = "voice"
	} else if msg.Photo != nil {
		fileType = "photo"
	} else if msg.Video != nil || msg.VideoNote != nil {
		fileType = "video"
	} else if ext == ".pdf" || ext == ".doc" || ext == ".docx" {
		fileType = "scan"
	}

	// 1. Prepare Directory: data/media/{patientID}
	baseDir := os.Getenv("DATA_DIR")
	if baseDir == "" {
		baseDir = "data"
	}
	mediaDir := filepath.Join(baseDir, "media", telegramID)
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		logging.Errorf("Failed to create media directory: %v", err)
		return c.Send("❌ Ошибка сервера (mkdir).")
	}

	// 2. Save File
	filePath := filepath.Join(mediaDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		logging.Errorf("Failed to create file: %v", err)
		return c.Send("❌ Ошибка сервера (create).")
	}

	if _, err := io.Copy(dst, fileReader); err != nil {
		dst.Close()
		logging.Errorf("Failed to save file content: %v", err)
		return c.Send("❌ Ошибка сервера (copy).")
	}
	dst.Close()

	// 3. Save Metadata to DB
	mediaID := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileName)

	telegramFileID := ""
	if msg.Document != nil {
		telegramFileID = msg.Document.FileID
	} else if msg.Photo != nil {
		telegramFileID = msg.Photo.FileID
	} else if msg.Voice != nil {
		telegramFileID = msg.Voice.FileID
	}

	// Store path relative to DATA_DIR for portability
	// baseDir is "data" or getenv("DATA_DIR")
	// mediaDir is baseDir/media/telegramID
	// filePath is baseDir/media/telegramID/fileName
	// We want to store "media/telegramID/fileName"
	relativePath := filepath.Join("media", telegramID, fileName)

	media := domain.PatientMedia{
		ID:             mediaID,
		PatientID:      telegramID,
		FileType:       fileType,
		FilePath:       relativePath, // Storing relative path
		TelegramFileID: telegramFileID,
		CreatedAt:      time.Now(),
	}

	if err := h.repository.SaveMedia(media); err != nil {
		logging.Errorf("Failed to save media metadata: %v", err)
		return c.Send("❌ Ошибка при сохранении метаданных.")
	}

	if statusMsg != nil {
		if err := c.Bot().Delete(statusMsg); err != nil {
			logging.Warnf("Failed to delete status message: %v", err)
		}
	}

	// Special handling for voice: Transcribe and save as Draft
	if voice := msg.Voice; voice != nil {
		transMsg, _ := c.Bot().Send(c.Recipient(), "📝 Расшифровываю ваше аудио-сообщение...")

		// We need a fresh reader or the content of the file
		fileReader, _ = c.Bot().File(&telebot.File{FileID: fileID})
		transcript, err := h.transcriptionService.Transcribe(context.Background(), fileReader, fileName)

		if transMsg != nil {
			_ = c.Bot().Delete(transMsg)
		}

		if err == nil && transcript != "" {
			// Save to media record as a draft
			media.Transcript = transcript
			media.Status = "pending"
			_ = h.repository.SaveMedia(media)

			// Notify Admins
			reviewMsg := h.presenter.FormatDraftNotification(patient.Name, transcript)
			
			// Inline buttons for quick action in the bot
			selector := &telebot.ReplyMarkup{}
			btnApprove := selector.Data("✅ В карту", "approve_draft", mediaID)
			btnDiscard := selector.Data("🗑️ Удалить", "discard_draft", mediaID)
			btnOpenTWA := selector.WebApp("📱 Открыть TWA", &telebot.WebApp{URL: h.GenerateWebAppURL(patient.TelegramID)})
			
			selector.Inline(
				selector.Row(btnApprove, btnDiscard),
				selector.Row(btnOpenTWA),
			)

			for _, adminIDStr := range h.adminIDs {
				adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
				h.BotNotify(c.Bot(), adminID, reviewMsg, selector)
			}
			
			return c.Send("✅ Сообщение получено. Терапевт скоро его изучит.")
		}
	} else {
		// Non-voice files
		if err := c.Send(fmt.Sprintf("✅ Файл '%s' успешно сохранен в вашу медицинскую карту!", fileName)); err != nil {
			logging.Warnf("Failed to send file saved message: %v", err)
		}
	}

	// Notify admins with HTML to avoid parsing errors with underscores in filenames
	details := map[string]string{
		"Пациент": html.EscapeString(patient.Name),
		"ID":      html.EscapeString(telegramID),
		"Файл":    fmt.Sprintf("<code>%s</code>", html.EscapeString(fileName)),
		"Размер":  fmt.Sprintf("%.2f MB", float64(fileSize)/(1024*1024)),
	}
	notification := h.presenter.FormatNotification("Новый файл в мед-карте", details)

	// Add link to med-card and Reply button
	selector := &telebot.ReplyMarkup{}
	btnReply := selector.Data("✍️ Ответить", "admin_reply", telegramID)
	selector.Inline(selector.Row(btnReply))

	for _, adminIDStr := range h.adminIDs {
		adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
		h.BotNotify(c.Bot(), adminID, notification, selector)
	}

	return nil
}

func (h *BookingHandler) HandleApproveDraft(c telebot.Context) error {
	data := strings.TrimPrefix(c.Callback().Data, "approve_draft|")
	parts := strings.Split(data, "|")
	if len(parts) < 1 {
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка данных"})
	}
	mediaID := parts[0]

	media, err := h.repository.GetMediaByID(mediaID)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Запись не найдена"})
	}

	// 1. Update status to approved
	err = h.repository.UpdateMediaStatus(mediaID, "approved", media.Transcript)
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка БД"})
	}

	// 2. Append to patient's clinical notes
	patient, err := h.repository.GetPatient(media.PatientID)
	if err == nil {
		newNotes := patient.TherapistNotes
		if newNotes != "" {
			newNotes += "\n\n"
		}
		timestamp := media.CreatedAt.Format("02.01.2006 15:04")
		newNotes += fmt.Sprintf("**Запись от %s:**\n%s", timestamp, media.Transcript)
		_ = h.repository.UpdatePatientProfile(media.PatientID, patient.Name, newNotes)
	}

	return c.Edit("✅ <b>ДОБАВЛЕНО В КАРТУ</b>\n\n" + media.Transcript, telebot.ModeHTML)
}

func (h *BookingHandler) HandleDiscardDraft(c telebot.Context) error {
	data := strings.TrimPrefix(c.Callback().Data, "discard_draft|")
	parts := strings.Split(data, "|")
	if len(parts) < 1 {
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка данных"})
	}
	mediaID := parts[0]

	media, err := h.repository.GetMediaByID(mediaID)
	if err == nil {
		_ = h.repository.UpdateMediaStatus(mediaID, "discarded", media.Transcript)
	} else {
		_ = h.repository.UpdateMediaStatus(mediaID, "discarded", "")
	}

	return c.Edit("🗑 <b>ЧЕРНОВИК УДАЛЕН</b>", telebot.ModeHTML)
}

