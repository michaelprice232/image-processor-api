# image-processor-api

Go REST based microservice for processing image files received from AWS S3 object creation events via AWS EventBridge. Uses [AWS Rekognition](https://docs.aws.amazon.com/rekognition/latest/dg/what-is.html) to validate with 80% confidence that only a single person appears in the image and that the person is not wearing sunglasses. For example, the service could be utilised as part of a validation workflow when a user updates their profile picture via an upstream webapp microservice.  

On successful image validation an event is posted to the `success-validate-profile-image-v1` Kafka topic. On failure a message is posted to the `failed-validate-profile-image-v1` topic. These are designed to be consumed by downstream services in an async manner.

The microservice exposes 2 endpoints:

1. `POST /validate` - consumes an S3 object creation event and validate the image file in the referenced in the payload. See `/testdata` for example event 
2. `GET /health` - health endpoint designed to be used by K8s health probes

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

## Sample AWS Architecture

![Sample AWS Architecture](/images/architecture.jpg)


## todo

- [] write unit tests
- [] write integration tests
- [] write Terraform for deploying into AWS
- [] configure /health probe to account for Kafka failures
- [] instrument app to gather key metrics
