package main

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
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

	log.Println("starting server in :8081")

	http.HandleFunc("/upload", uploadFileAndSendToServer)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func uploadFileAndSendToServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		http.Error(w, "Error creating form file", http.StatusInternalServerError)
		return
	}

	if _, err = part.Write(fileBytes); err != nil {
		http.Error(w, "Error writing file to form", http.StatusInternalServerError)
		return
	}

	if err = writer.Close(); err != nil {
		http.Error(w, "Error closing writer", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/upload", &buf)
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(http.StatusOK)
}

// func getVideoMetadata(filepath string) (*FFprobeMetadata, error) {
// 	cmd := exec.Command(
// 		"ffprobe",
// 		"-v", "error",
// 		"-show_format",
// 		"-show_streams",
// 		"-print_format", "json",
// 		filepath,
// 	)

// 	output, err := cmd.Output()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var metadata FFprobeMetadata
// 	if err := json.Unmarshal(output, &metadata); err != nil {
// 		return nil, fmt.Errorf("ошибка парсинга JSON ffprobe: %w", err)
// 	}

// 	return &metadata, nil
// }

// func compressVideoWithPresets(filepath string) error {
// 	start := time.Now()
// 	cmd := exec.Command(
// 		"ffmpeg",
// 		"-y",
// 		"-i", filepath,
// 		"-c:v", "libx264",
// 		"-preset", "ultrafast",
// 		"-crf", "28",
// 		OutputFile,
// 	)

// 	cmd.Stderr = os.Stderr
// 	cmd.Stdout = os.Stdout
// 	if err := cmd.Run(); err != nil {
// 		return fmt.Errorf("ошибка конвертации в MP4 с прессетом: %w", err)
// 	}
// 	duration := time.Since(start)

// 	log.Println("Time for compress video with presets:", duration)
// 	return nil
// }

// func compressVideo(filepath string) error {
// 	now := time.Now()
// 	cmd := exec.Command(
// 		"ffmpeg",
// 		"-y",
// 		"-i", filepath,
// 		"-c:v", "h264_videotoolbox",
// 		"-preset", "ultrafast",
// 		"-crf", "28",
// 		"-b:v", "3000k",
// 		"-c:a", "aac",
// 		"-b:a", "128k",
// 		OutputFile,
// 	)

// 	cmd.Stderr = os.Stderr
// 	cmd.Stdout = os.Stdout
// 	if err := cmd.Run(); err != nil {
// 		return fmt.Errorf("ошибка конвертации в MP4: %w", err)
// 	}

// 	duration := time.Since(now)

// 	log.Println("Time for compress video:", duration)

// 	return nil
// }

// func modifyVideoParam(filepath string) error {
// 	cmd := exec.Command(
// 		"ffmpeg",
// 		"-i", filepath,
// 		// "-c:v", Codec,
// 		"-b:v", Bitrate,
// 		// "-vf", "scale="+Resolution,
// 		"-y",
// 		OutputFile,
// 	)

// 	output, err := cmd.Output()
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println(string(output))
// 	return nil
// }

// func convertToWebm(filepath string) error {
// 	cmd := exec.Command("ffmpeg",
// 		"-i", filepath,
// 		"-c:v", "libvpx-vp9",
// 		"-b:v", "1M",
// 		"-c:a", "libopus",
// 		"-b:a", "128k",
// 		"-y",
// 		OutputFileWebm,
// 	)
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return fmt.Errorf("ошибка конвертации в WebM: %w. ffmpeg output:\n%s", err, output)
// 	}
// 	return nil
// }
