package ytVideoMaker

import (
	"bytes"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

var (
	ytClient   *youtube.Client
	clientOnce sync.Once
)

func InitClient(proxyUrl string) error {
	var initErr error
	clientOnce.Do(func() {
		proxy, err := url.Parse(proxyUrl)
		if err != nil {
			initErr = fmt.Errorf("failed to parse proxy URL: %w", err)
			return
		}

		httpCustomClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
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

// Reeturn Title, destinarion folder, error
func DownloadVideoAndAudio(videoUrl string, destFolderPrefix string, start int, end int) (string, string, error) {
	video, err := ytClient.GetVideo(videoUrl)
	if err != nil {
		return "", "", err
	}

	destFolder := fmt.Sprintf("%s_%s_%d_%d", destFolderPrefix, video.ID, start, end)
	err = os.Mkdir(destFolder, os.ModePerm)
	if err != nil {
		panic(err)
	}

	// save audio.mp4
	formats := video.Formats.WithAudioChannels()
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
	defer os.Remove(fmt.Sprintf("%s/video.mp4", destFolder))
	defer os.Remove(fmt.Sprintf("%s/audio.mp4", destFolder))
	wg.Wait()

	if end > int(video.Duration.Seconds()) {
		end = int(video.Duration.Seconds())
	}

	args := []string{
		"-i",
		fmt.Sprintf("%s/video.mp4", destFolder),
		"-i",
		fmt.Sprintf("%s/audio.mp4", destFolder),
		"-vf",
		"crop=ih:ih,scale=512:512",
		"-ss",
		strconv.Itoa(start),
		"-t",
		strconv.Itoa(end),
		"-c:v",
		"libx264",
		"-c:a",
		"aac",
		"-preset",
		"ultrafast",
		fmt.Sprintf("%s/output.mp4", destFolder),
	}

	cmd := exec.Command(
		"ffmpeg",
		args...,
	)

	// Захватываем stderr для диагностики ошибок
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("ffmpeg merge failed: %w, stderr: %s", err, stderr.String())
	}
	return video.Title, destFolder, nil
}
