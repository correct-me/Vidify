package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

const InputFile = "media/Енисей_барберинг.mp4"
const OutputFile = "media/Енисей_барберинг_new.mp4"
const OutputFileWebm = "media/Енисей_барберинг_new.webm"

var Codec = ""
var Bitrate = "100000"
var Resolution = ""

type FFprobeMetadata struct {
	Streams []struct {
		Index   int    `json:"index"`
		Bitrate string `json:"bit_rate"`
	} `json:"streams"`
	Format struct {
		Filename string `json:"filename"`
		Bitrate  string `json:"bit_rate"`
	} `json:"format"`
}

func main() {
	// metadataOld, err := getVideoMetadata(InputFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	if err := compressVideo(InputFile); err != nil {
		log.Fatal(err)
	}
	log.Println("video compressed successfully")

	if err := compressVideoWithPresets(InputFile); err != nil {
		log.Fatal(err)
	}
	log.Println("video compressed successfully")

	// if err := convertToWebm(OutputFile); err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("video converted sucessfully")

	// if err := modifyVideoParam(InputFile); err != nil {
	// 	log.Fatal(err)
	// }

	// metadataNew, err := getVideoMetadata(OutputFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(metadataNew.Format.Bitrate)
}

func getVideoMetadata(filepath string) (*FFprobeMetadata, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_format",
		"-show_streams",
		"-print_format", "json",
		filepath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var metadata FFprobeMetadata
	if err := json.Unmarshal(output, &metadata); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON ffprobe: %w", err)
	}

	return &metadata, nil
}

func compressVideoWithPresets(filepath string) error {
	start := time.Now()
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", filepath,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "28",
		OutputFile,
	)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка конвертации в MP4 с прессетом: %w", err)
	}
	duration := time.Since(start)

	log.Println("Time for compress video with presets:", duration)
	return nil
}

func compressVideo(filepath string) error {
	now := time.Now()
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", filepath,
		"-c:v", "h264_videotoolbox",
		"-preset", "ultrafast",
		"-crf", "28",
		"-b:v", "3000k",
		"-c:a", "aac",
		"-b:a", "128k",
		OutputFile,
	)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ошибка конвертации в MP4: %w", err)
	}

	duration := time.Since(now)

	log.Println("Time for compress video:", duration)

	return nil
}

func modifyVideoParam(filepath string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", filepath,
		// "-c:v", Codec,
		"-b:v", Bitrate,
		// "-vf", "scale="+Resolution,
		"-y",
		OutputFile,
	)

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

func convertToWebm(filepath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", filepath,
		"-c:v", "libvpx-vp9",
		"-b:v", "1M",
		"-c:a", "libopus",
		"-b:a", "128k",
		"-y",
		OutputFileWebm,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка конвертации в WebM: %w. ffmpeg output:\n%s", err, output)
	}
	return nil
}
