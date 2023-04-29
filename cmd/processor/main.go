package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/michaelprice232/image-processor-api/internal/validate-profile"
)

func main() {
	bucketName := "mike-dev-80239839823"
	image := "single-face.jpg"
	//image := "multiple-faces.jpg"
	//image := "cartoon-face.jpg"
	//image := "no-face.png"
	//image := "non-existent-file.png"
	//image := "with-sunglasses.jpg"

	client, err := validate_profile.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	err = client.ProcessImage(bucketName, image)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Procesed successfully")
}
