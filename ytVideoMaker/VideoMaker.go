package ytVideoMaker

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"ytubecircles/config"
)

var (
	ytClient   *youtube.Client
	clientOnce sync.Once
)

func InitClient(proxyUrl string) error {
	var initErr error
	clientOnce.Do(func() {
		httpCustomClient := http.DefaultClient

		if proxyUrl != "" {
			proxy, err := url.Parse(proxyUrl)
			if err != nil {
				initErr = fmt.Errorf("failed to parse proxy URL: %w", err)
				return
			}
			httpCustomClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
		}
		ytClient = &youtube.Client{HTTPClient: httpCustomClient}
	})
	return initErr
}

func downloadFormat(video *youtube.Video, format *youtube.Format, destination string, wg *sync.WaitGroup) error {
	defer wg.Done()
	stream, _, err := ytClient.GetStream(video, format)
	if err != nil {
		log.Printf("Failed to get stream %s: %v", destination, err)
		return err
	}
	defer stream.Close()

	file, err := os.Create(destination)
	if err != nil {
		log.Printf("Failed to create file %s: %v", destination, err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		log.Printf("Failed to copy stream to file %s: %v", destination, err)
		return err
	}
	return nil
}

func encodeVideo(videoFormat *youtube.Format, videoFile string, audioFile string, outFile string, start string, end string) error {
	var cropMode string
	if videoFormat.Height > videoFormat.Width {
		cropMode = "crop=iw:iw,scale=512:512"
	} else {
		cropMode = "crop=ih:ih,scale=512:512"
	}

	args := []string{
		"-i",
		videoFile,
		"-i",
		audioFile,
		"-vf",
		cropMode,
		"-ss",
		start,
		"-t",
		end,
		"-c:v",
		"libx264",
		"-c:a",
		"aac",
		"-preset",
		"ultrafast",
		outFile,
	}

	cmd := exec.Command(
		config.Config.FFMPEGBin,
		args...,
	)

	log.Printf("Running: %s %s", cmd.Path, strings.Join(cmd.Args, " "))

	// Захватываем stderr для диагностики ошибок
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg merge failed: %w, stderr: %s", err, stderr.String())
	}
	return nil
}

// DownloadVideoAndAudio загружает видео и аудио с YouTube, обрабатывает их и возвращает заголовок видео и путь к итоговому файлу.
func DownloadVideoAndAudio(videoUrl string, destFolderPrefix string, start float64, duration float64) (string, string, error) {
	video, err := ytClient.GetVideo(videoUrl)
	if err != nil {
		return "", "", err
	}

	destFolder := fmt.Sprintf("%s_%s_%f_%f", destFolderPrefix, video.ID, start, duration)
	err = os.Mkdir(destFolder, os.ModePerm)
	if err != nil {
		rem_err := os.RemoveAll(destFolder)
		if rem_err != nil {
			return "", "", fmt.Errorf("failed to remove folder %s: %v", destFolder, rem_err)
		}
		err = os.Mkdir(destFolder, os.ModePerm)
		if err != nil {
			return "", "", err
		}
	}

	// save audio.mp4
	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return "", "", errors.New("no suitable audio formats found")
	}
	bestAudioFormat := formats[0]
	for _, format := range formats {
		if format.AverageBitrate > bestAudioFormat.AverageBitrate {
			bestAudioFormat = format
		}
	}
	log.Printf("Best audio format: %#v\n", bestAudioFormat)

	formats = video.Formats.Select(func(format youtube.Format) bool {
		return format.Width != 0
	})
	if len(formats) == 0 {
		return "", "", errors.New("no suitable video formats found")
	}
	bestVideoFormat := formats[0]
	for _, format := range formats {
		if format.Height < 512 && (bestVideoFormat.Height < format.Height || bestVideoFormat.Height > 512) {
			bestVideoFormat = format
		}
	}
	log.Printf("Best video format: %#v\n", bestVideoFormat)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go downloadFormat(video, &bestAudioFormat, destFolder+"/audio.mp4", wg)
	go downloadFormat(video, &bestVideoFormat, destFolder+"/video.mp4", wg)
	wg.Wait()

	if duration > video.Duration.Seconds()-start {
		duration = video.Duration.Seconds() - start
	}

	err = encodeVideo(
		&bestVideoFormat,
		fmt.Sprintf("%s/video.mp4", destFolder),
		fmt.Sprintf("%s/audio.mp4", destFolder),
		fmt.Sprintf("%s/output.mp4", destFolder),
		strconv.FormatFloat(start, 'f', 3, 64),
		strconv.FormatFloat(duration, 'f', 3, 64),
	)
	if err != nil {
		return "", "", err
	}

	return video.Title, destFolder, nil
}
