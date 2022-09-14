# pi.delivery

This repository has source code that runs [pi.delivery](https://pi.delivery), which is
Google Cloud DevRel's demo site to celebrate Pi Day.

This is not an official google product.

# Server

The server is written in Go and runs in [Cloud Functions](https://cloud.google.com/functions/docs/2nd-gen/overview).

The entry point is the Get function in [functions.go](functions.go).

## Infrastructure

![Server architecture diagram. There's a Cloud Load Balancer in the front that redirects requests to Cloud Function instances in us-central1, europe-west1, asia-northeast1 regions. The functions connect to Cloud Storage in the US multi-region. Logging and Monitoring are used for monitoring. Cloud DNS is used for DNS resolutions.](docs/server-diagram.svg)

Cloud configurations are managed with [Terraform](https://www.terraform.io/). The configuration file is [main.tf](main.tf), and uses the [Google Cloud Platform Provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs).

Most of the configurations are done by terraform with a few exceptions.

```bash
terraform init
terraform apply
```

## Functions (deploy with gcloud)

Set environment variables (we use Bash extensions here).

```bash
export PROJECT=piaas-gcp
export REGIONS=(us-central1 europe-west1 asia-northeast1)
export STAGE_BUCKET=piaas-gcp-gcf-staging
export GCF_API_SA=sa-functions-api@piaas-gcp.iam.gserviceaccount.com
```

### Staging

```bash
gcloud beta functions deploy api-pi-staging --gen2 --runtime go116 --trigger-http --entry-point Get --source . \
  --stage-bucket=$STAGE_BUCKET --ingress-settings=internal-and-gclb --region=$REGIONS[1] \
  --allow-unauthenticated --service-account=$GCF_API_SA --project=$PROJECT
```

The staging API is accessible via `api.staging.pi.delivery`. e.g.

```bash
curl -v https://api.staging.pi.delivery/v1/pi
```

### Production

The whole setup is just experimental but we call it "production".

```bash
for R in $REGIONS; \
  gcloud beta functions deploy api-pi --gen2 --runtime go116 --trigger-http --entry-point Get --source . \
    --stage-bucket=$STAGE_BUCKET --ingress-settings=internal-and-gclb --region=$R \
    --allow-unauthenticated --service-account=$GCF_API_SA \
    --min-instances=1  --project=$PROJECT
```

The production API is accessible via `api.pi.delivery`.

```bash
curl -v https://api.pi.delivery/v1/pi
```

## Utility commands

There are several utility commands under [cmd/](cmd/).

### dtob

The dtob command is a small tool to generate binary data used in some of the test cases.
It takes a pi digit text as input and generates hex values that you can paste in a Go source file.
For example, [pkg/unpack/reader_test.go](./pkg/unpack/reader_test.go) have test cases that check decoding of compressed
digits that use outputs of this program.

```bash
tail -c +3 pi.txt | head -c 100 | go run cmd/dtob/main.go -b 30
```

The output should look like this:

```go
0x60, 0xe2, 0x3e, 0xb8, 0xae, 0x61, 0xa6, 0x13,
0x00, 0x0f, 0x58, 0xf3, 0x84, 0x66, 0xef, 0x56,
// Block Boundary
0x17, 0x3f, 0x65, 0x1a, 0x21, 0x09, 0xca, 0x45,
0x00, 0x60, 0x5b, 0x4a, 0x96, 0x06, 0x14, 0x08,
// Block Boundary
0x09, 0xfb, 0xd6, 0x59, 0x35, 0x00, 0x33, 0x52,
0x00, 0xe9, 0xe5, 0x0f, 0x83, 0xb7, 0xdf, 0x88,
// Block Boundary
0x00, 0xe6, 0xc6, 0x3d, 0x9b, 0x70, 0x7a, 0x2f,
// Block Boundary
```

### extact

This is a command line version of the API that uses the same code to fetch and parse ycd files.
The major difference is that you can fetch as many digits as you'd like with this program.

```bash
go run ./cmd/extract -s 42 -n 2000
```

### indexer

The indexer program is used to generate the index files that the API needs to determine which object to fetch.
The generated file needs to be in [gen/index/](./gen/index/).

```bash
go run ./cmd/indexer --bucket pi50t >  gen/index/index.go
```

### rest

This is a command line emulator of the Functions API.
Check out [functions-framework-go](https://github.com/GoogleCloudPlatform/functions-framework-go) to learn more about the framework.

# Frontend

The frontend is developed with [Jekyll](https://jekyllrb.com/) and [React](https://reactjs.org/).

- [src/](./src/) - source code of the web browser demos
- [jekyll/](./jekyll/) - static content for Jekyll
- [third_party/aviator-jekyll-template](./third_party/aviator-jekyll-template/) - Jekyll template ([repo](https://github.com/CloudCannon/aviator-jekyll-template))

## Prerequisites

You need [Yarn](https://yarnpkg.com/) and [Ruby](https://www.ruby-lang.org/en/) to compile the frontend.
We use Yarn 3 and Ruby 3.1.

To install dependencies, run

```bash
yarn
bundle
```

## Develop

There is a Yarn script to build and monitor changes for both Jekyll and Webpack and run a development server.

```bash
yarn serve
```

## Build

To build minimized files for deployment, run:

```bash
yarn prod
```

## Deploy

We use [Firebase Hosting](https://firebase.google.com/docs/hosting) to server the frontend files.

Firebase hosting is separately managed with the Firebase command.
To install the tool, run the following command:

```bash
npm install -g firebase-tools
firebase login
```

Make sure to build the latest production build before testing or deploying with Firebase.
The files will be generated in `dist` directory.

```bash
yarn prod
```

Test the configuration (such as [firebase.json](firebase.json)) locally by running:

```bash
export PROJECT=piaas-gcp
firebase emulators:start --project=${PROJECT}
```

You can also create preview channels to test the website on Firebase servers. `CHANNEL_ID` should be set to an identifier you want to use (e.g. preview)

```bash
export CHANNEL_ID=preview
firebase hosting:channel:deploy ${CHANNEL_ID} --project=${PROJECT}
```

After confirming the preview, run the following command to serve it live:

```bash
firebase hosting:clone ${PROJECT}:${CHANNEL_ID} ${PROJECT}:live
```

See the [official documentation](https://firebase.google.com/docs/hosting/test-preview-deploy) for more details on the testing and deployment process with Firebase.
