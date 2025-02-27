package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
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

	http.HandleFunc("/proxy", proxyHandler)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем файл из формы (ключ "file")
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения файла: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Сохраняем исходное видео во временный файл
	inputFile := "temp_input.mp4"
	outFile, err := os.Create(inputFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка создания временного файла: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(outFile, file); err != nil {
		outFile.Close()
		http.Error(w, fmt.Sprintf("Ошибка сохранения файла: %v", err), http.StatusInternalServerError)
		return
	}
	outFile.Close()

	// Сжимаем видео с помощью ffmpeg
	outputFile := "temp_output.mp4"
	cmd := exec.Command("ffmpeg", "-y", "-i", inputFile, "-c:v", "libx264", "-preset", "ultrafast", "-crf", "28", outputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка сжатия видео: %v", err), http.StatusInternalServerError)
		return
	}

	// Формирование multipart-запроса для отправки сжатого файла
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Создаем часть формы с ключом "file"
	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка создания части формы: %v", err), http.StatusInternalServerError)
		return
	}

	// Открываем сжатый файл и копируем его содержимое в multipart часть
	compFile, err := os.Open(outputFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка открытия сжатого файла: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(part, compFile); err != nil {
		compFile.Close()
		http.Error(w, fmt.Sprintf("Ошибка записи файла в форму: %v", err), http.StatusInternalServerError)
		return
	}
	compFile.Close()

	// Завершаем формирование multipart данных
	if err := writer.Close(); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка закрытия writer: %v", err), http.StatusInternalServerError)
		return
	}

	// Формируем новый POST-запрос для отправки на целевой сервис
	targetURL := "http://localhost:8080/upload"
	req, err := http.NewRequest("POST", targetURL, &buf)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка создания запроса: %v", err), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос на целевой сервис
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка отправки запроса: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Читаем ответ от сервиса и возвращаем его клиенту
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка отправки ответа клиенту: %v", err), http.StatusInternalServerError)
		return
	}

	// Удаляем временные файлы
	os.Remove(inputFile)
	os.Remove(outputFile)
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
