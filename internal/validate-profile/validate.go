package validate_profile

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	log "github.com/sirupsen/logrus"
)

const sunGlassesConfidenceLevelPercent float32 = 80.0
const realFaceConfidenceLevelPercent float32 = 80.0

// ProcessImage uses AWS Rekognition to validate an image file
// Validates that only a single face appears in the image, and it isn't wearing sunglasses
// Only jpeg, jpg and png are supported formats
// Returns nil if processed successfully with no errors
func (c *Client) ProcessImage(s3Bucket, s3FilePath string) error {
	log.Debugf("Processing bucket: %s, file: %s", s3Bucket, s3FilePath)

	if !validateFileExtension(s3FilePath) {
		return fmt.Errorf("error: only jpeg, jpg and png image formats are supported")
	}

	input := &rekognition.DetectFacesInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String(s3Bucket),
				Name:   aws.String(s3FilePath),
			},
		},
		Attributes: []types.Attribute{
			"ALL",
		},
	}

	resp, err := getFaces(context.TODO(), c.client, input)
	if err != nil {
		return fmt.Errorf("error: calling getFaces: %v", err)
	}

	if err = validateImage(resp.FaceDetails); err != nil {
		return err
	}

	return nil
}

// validateFileExtension confirms we are processing a supported extension. Returns false if invalid
func validateFileExtension(s3FilePath string) bool {
	fileExtension := filepath.Ext(s3FilePath)
	validExtension := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
	}
	if !validExtension[fileExtension] {
		return false
	}

	return true
}

// validateImage validates that we have 1 face within the image which isn't wearing glasses
func validateImage(faceDetails []types.FaceDetail) error {
	facesDetected := len(faceDetails)
	// ensure only 1 face appears in the image
	if facesDetected != 1 {
		return fmt.Errorf("error: number of faces found: %d. exactly 1 face needs to be detected", facesDetected)
	}

	for _, faceDetail := range faceDetails {
		// check for sunglasses
		if faceDetail.Sunglasses != nil && faceDetail.Sunglasses.Value && *faceDetail.Sunglasses.Confidence > sunGlassesConfidenceLevelPercent {
			return fmt.Errorf("error: sunglasses detected with > %v%% confidence", sunGlassesConfidenceLevelPercent)
		}

		// ensure we have a high degree of confidence that a face appears
		if *faceDetail.Confidence < realFaceConfidenceLevelPercent {
			return fmt.Errorf("error: less than %v%% condidence that a single face appears in this image", realFaceConfidenceLevelPercent)
		}
	}

	return nil
}

// RekognitionDetectFacesAPI defines the interface for the DetectFaces function
type rekognitionDetectFacesAPI interface {
	DetectFaces(ctx context.Context, params *rekognition.DetectFacesInput, optFns ...func(*rekognition.Options)) (*rekognition.DetectFacesOutput, error)
}

// getFaces is a wrapper around the rekognitionDetectFacesAPI interface to allow us to mock the API for unit tests
func getFaces(c context.Context, api rekognitionDetectFacesAPI, input *rekognition.DetectFacesInput) (*rekognition.DetectFacesOutput, error) {
	resp, err := api.DetectFaces(c, input)
	return resp, err
}
