# image-processor-api

Go REST based microservice for processing image files received from AWS S3 object creation events via AWS EventBridge. Uses [AWS Rekognition](https://docs.aws.amazon.com/rekognition/latest/dg/what-is.html) to validate with 80% confidence that exactly one person (i.e. not a cartoon person) appears in the image and that the person is not wearing sunglasses. This logic could be extended as needed. For example, the service could be utilised as part of a validation workflow when a user updates their profile picture via an upstream webapp microservice.  

On successful image validation an event is posted to the `success-validate-profile-image-v1` Kafka topic. On failure a message is posted to the `failed-validate-profile-image-v1` topic. These are designed to be consumed by downstream services in an async manner. The Kafka event schema is located in the `kafkaResponseEvent` struct in `internal/validate-profile/kakfa-producer.go`.

The microservice exposes 2 endpoints:

1. `POST /validate` - consumes an S3 object creation event (via EventBridge) and validates the image file referenced in the payload using AWS Rekognition. See `/testdata` for example event 
2. `GET /health` - health endpoint designed to be used by K8s health probes

The `/validate` route is designed to be exposed publicly so that it can act as an API destination in EventBridge. For that reason middleware has been added to require authentication via an API key. Required header:
```text
# Required auth HTTP header
Authorization: Bearer <api-key>
```

The `/validate` endpoint only supports the following image extensions: `jpeg, jpg, png` although typically the filtering would be done at the EventBridge rule level.



## Running locally via docker-compose

Pre-reqs:
- Docker
- Docker Compose
- AWS profile configured locally with access to the S3 and Rekognition services
- 4GB+ of RAM available to Docker, to run the Kafka stack and the app

```shell
export AWS_PROFILE=<my-profile>
export API_KEY=<API_KEY>          # <== API bearer token to use for authentication from clients
docker-compose up -d
```

### Test Endpoints using curl/Postman
Typically, the app is driven by EventBridge events but for local testing you can upload an image file to S3 and then pass a handcraft event to the endpoint:
```shell
## Health endpoint
curl -v localhost:3000/health

## Validate image endpoint
# 1. Upload an image file (jpg/jpeg/png) to an S3 bucket (Rekognition pulls from here)
# 2. Update the ./testdata/s3-object-create-event.json file to reference the S3 bucket name and file key 
# 3. Run:
curl -v -X POST -H "Content-Type: application/json" -H "Authorization: Bearer test" \
  -d @./testdata/s3-object-create-event.json \
  localhost:3000/validate
```
### Viewing Kafka messages
```shell
# Script to view the messages in a Kafka topic (goes from the beginning)
./scripts/consume-messages.sh success-validate-profile-image-v1 # valid images
./scripts/consume-messages.sh failed-validate-profile-image-v1  # invalid images
```


## Sample AWS Architecture

![Sample AWS Architecture](/images/architecture.jpg)


## todo

- [ ] write unit tests
- [ ] write integration tests
- [ ] create CI pipeline
- [ ] write Terraform for deploying into AWS
- [ ] configure /health probe to account for Kafka failures
- [ ] instrument app to gather key metrics
