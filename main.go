package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

const InputFile = "media/8582174042837694306.mp4"
const OutputFile = "media/8582174042837694306_new.mp4"

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
	metadataOld, err := getVideoMetadata(InputFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := modifyVideoParam(InputFile); err != nil {
		log.Fatal(err)
	}

	metadataNew, err := getVideoMetadata(OutputFile)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(metadataOld.Format.Bitrate)
	fmt.Println(metadataNew.Format.Bitrate)
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
