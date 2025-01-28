package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	inputFile := "media/8582174042837694306.mp4"

	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_format",
		"-show_streams",
		"-print_format", "json",
		inputFile,
	)

	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))

}
