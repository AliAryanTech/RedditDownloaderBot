package bot

import (
	"github.com/HirbodBehnam/RedditDownloaderBot/reddit"
	"github.com/HirbodBehnam/RedditDownloaderBot/util"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strings"
	"time"
)

// handleGifUpload downloads a gif and then uploads it to Telegram
func handleGifUpload(gifUrl, title, thumbnailUrl string, chatID int64) {
	// Inform the user we are doing some shit
	stopReportChannel := statusReporter(chatID, "upload_video")
	defer close(stopReportChannel)
	// Download the gif
	tmpFile, err := reddit.DownloadGif(gifUrl)
	if err != nil {
		log.Println("Cannot download file", gifUrl, ":", err)
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Cannot download file.\nHere is the link to file: "+gifUrl))
		return
	}
	defer func() { // Cleanup
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()
	// Upload the gif
	// Check file size
	if !util.CheckFileSize(tmpFile.Name(), RegularMaxUploadSize) {
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "This file is too big to upload it on telegram!\nHere is the link to file: "+gifUrl))
		return
	}
	// Check thumbnail
	var tmpThumbnailFile *os.File = nil
	if !util.CheckFileSize(tmpFile.Name(), NoThumbnailNeededSize) && thumbnailUrl != "" {
		tmpThumbnailFile, err = reddit.DownloadThumbnail(thumbnailUrl)
		if err == nil {
			defer func() {
				_ = tmpThumbnailFile.Close()
				_ = os.Remove(tmpThumbnailFile.Name())
			}()
		}
	}
	// Upload it
	msg := tgbotapi.NewAnimation(chatID, tmpFile.Name())
	msg.Caption = title
	if tmpThumbnailFile != nil {
		msg.Thumb = tmpThumbnailFile.Name()
	}
	_, err = bot.Send(msg)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Cannot upload file.\nHere is the link to file: "+gifUrl))
		log.Println("Cannot upload file:", err)
		return
	}
}

// handleVideoUpload downloads a video and then uploads it to Telegram
func handleVideoUpload(vidUrl, title, thumbnailUrl string, chatID int64) {
	// Inform the user we are doing some shit
	stopReportChannel := statusReporter(chatID, "upload_video")
	defer close(stopReportChannel)
	// Download the gif
	audioUrl, tmpFile, err := reddit.DownloadVideo(vidUrl)
	if err != nil {
		log.Println("Cannot download file", vidUrl, ":", err)
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Can't download file.\n"+generateVideoUrlsMessage(vidUrl, audioUrl)))
		return
	}
	defer func() { // Cleanup
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()
	// Check file size
	if !util.CheckFileSize(tmpFile.Name(), RegularMaxUploadSize) {
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "This file is too big to upload it on telegram!\n"+generateVideoUrlsMessage(vidUrl, audioUrl)))
		return
	}
	// Check thumbnail
	var tmpThumbnailFile *os.File = nil
	if !util.CheckFileSize(tmpFile.Name(), NoThumbnailNeededSize) && thumbnailUrl != "" {
		tmpThumbnailFile, err = reddit.DownloadThumbnail(thumbnailUrl)
		if err == nil {
			defer func() {
				_ = tmpThumbnailFile.Close()
				_ = os.Remove(tmpThumbnailFile.Name())
			}()
		}
	}
	// Upload it
	msg := tgbotapi.NewVideo(chatID, tmpFile.Name())
	msg.Caption = title
	if tmpThumbnailFile != nil {
		msg.Thumb = tmpThumbnailFile.Name()
	}
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("Cannot upload file:", err)
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Cannot upload file.\n"+generateVideoUrlsMessage(vidUrl, audioUrl)))
		return
	}
}

// handleVideoUpload downloads a photo and then uploads it to Telegram
func handlePhotoUpload(photoUrl, title, thumbnailUrl string, chatID int64, asPhoto bool) {
	// Inform the user we are doing some shit
	var stopReportChannel chan struct{}
	if asPhoto {
		stopReportChannel = statusReporter(chatID, "upload_photo")
	} else {
		stopReportChannel = statusReporter(chatID, "upload_document")
	}
	defer close(stopReportChannel)
	// Download the gif
	tmpFile, err := reddit.DownloadPhoto(photoUrl)
	if err != nil {
		log.Println("Cannot download file", photoUrl, ":", err)
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Cannot download file.\nHere is the link to file: "+photoUrl))
		return
	}
	defer func() { // Cleanup
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()
	// Check filesize
	if asPhoto {
		asPhoto = util.CheckFileSize(tmpFile.Name(), PhotoMaxUploadSize) // send photo as file if it is larger than 10MB
	}
	if !util.CheckFileSize(tmpFile.Name(), RegularMaxUploadSize) {
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "This file is too big to upload it on telegram!\nHere is the link to image: "+photoUrl))
		return
	}
	// Download thumbnail
	var tmpThumbnailFile *os.File = nil
	if !util.CheckFileSize(tmpFile.Name(), NoThumbnailNeededSize) && thumbnailUrl != "" {
		tmpThumbnailFile, err = reddit.DownloadThumbnail(thumbnailUrl)
		if err == nil {
			defer func() {
				_ = tmpThumbnailFile.Close()
				_ = os.Remove(tmpThumbnailFile.Name())
			}()
		}
	}
	// Upload
	var msg tgbotapi.Chattable
	if asPhoto {
		photo := tgbotapi.NewPhoto(chatID, tmpFile.Name())
		photo.Caption = title
		if thumbnailUrl != "" {
			photo.Thumb = tmpThumbnailFile.Name()
		}
		msg = photo
	} else {
		photo := tgbotapi.NewDocument(chatID, tmpFile.Name())
		photo.Caption = title
		if thumbnailUrl != "" {
			photo.Thumb = tmpThumbnailFile.Name()
		}
		msg = photo
	}
	_, err = bot.Send(msg)
	if err != nil {
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Cannot upload file.\nHere is the link to image: "+photoUrl))
		log.Println("Cannot upload file:", err)
		return
	}
}

// handleAlbumUpload uploads an album to Telegram
func handleAlbumUpload(album reddit.FetchResultAlbum, chatID int64) {
	// Report status
	stopReportChannel := statusReporter(chatID, "upload_photo")
	defer close(stopReportChannel)
	// Download each file of album
	var err error
	filePaths := make([]*os.File, 0, len(album.Album))
	defer func() { // cleanup
		for _, f := range filePaths {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}
	}()
	fileConfigs := make([]interface{}, 0, len(album.Album))
	for _, media := range album.Album {
		var tmpFile *os.File
		switch media.Type {
		case reddit.FetchResultMediaTypePhoto:
			tmpFile, err = reddit.DownloadPhoto(media.Link)
			if err == nil {
				f := tgbotapi.NewInputMediaPhoto(tmpFile.Name())
				f.Caption = media.Caption
				fileConfigs = append(fileConfigs, f)
			}
		case reddit.FetchResultMediaTypeGif:
			tmpFile, err = reddit.DownloadGif(media.Link)
			if err == nil {
				f := tgbotapi.NewInputMediaVideo(tmpFile.Name()) // not sure why...
				f.Caption = media.Caption
				fileConfigs = append(fileConfigs, f)
			}
		case reddit.FetchResultMediaTypeVideo:
			_, tmpFile, err = reddit.DownloadVideo(media.Link)
			if err == nil {
				f := tgbotapi.NewInputMediaVideo(tmpFile.Name())
				f.Caption = media.Caption
				fileConfigs = append(fileConfigs, f)
			}
		}
		if err != nil {
			log.Println("cannot download media of gallery:", err)
			continue
		}
		filePaths = append(filePaths, tmpFile)
	}
	// Now upload 10 of them at once
	i := 0
	for ; i < len(fileConfigs)/10; i++ {
		_, err = bot.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, fileConfigs[i*10:(i+1)*10]))
		if err != nil {
			log.Println("Cannot upload gallery:", err)
		}
	}
	err = nil
	fileConfigs = fileConfigs[i*10:]
	if len(fileConfigs) == 1 {
		switch f := fileConfigs[0].(type) {
		case tgbotapi.InputMediaPhoto:
			_, err = bot.Send(tgbotapi.NewPhoto(chatID, f.Media))
		case tgbotapi.InputMediaVideo:
			_, err = bot.Send(tgbotapi.NewVideo(chatID, f.Media))
		}
	} else if len(fileConfigs) > 1 {
		_, err = bot.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, fileConfigs))
	}
	if err != nil {
		log.Println("cannot upload gallery:", err)
	}
}

// statusReporter starts reporting for uploading a thing in telegram
// This function returns a channel which a message must be sent to it when reporting must be stopped
// You can also close the channel to stop the reporter
func statusReporter(chatID int64, action string) chan struct{} {
	doneChan := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(time.Second * 5) // we have to send it each 5 seconds
		_, _ = bot.Send(tgbotapi.NewChatAction(chatID, action))
		for {
			select {
			case <-ticker.C:
				_, _ = bot.Send(tgbotapi.NewChatAction(chatID, action))
			case <-doneChan:
				ticker.Stop()
				return
			}
		}
	}()
	return doneChan
}

// generateVideoUrlsMessage generates a text message which it can be used to give the user
// the requested video and audio URL
func generateVideoUrlsMessage(videoUrl, audioUrl string) string {
	var sb strings.Builder
	sb.Grow(150)
	sb.WriteString("Here is the link to video file: ")
	sb.WriteString(videoUrl)
	if audioUrl != "" {
		sb.WriteString("\nHere is the link to audio file: ")
		sb.WriteString(audioUrl)
	}
	return sb.String()
}
