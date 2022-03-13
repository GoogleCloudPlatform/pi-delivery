---
title: Serving the API
position_number: 1
---

We developed the REST API in Go and running the program in [Cloud Functions (2nd gen)](https://cloud.google.com/functions/docs/2nd-gen/overview).

The architecture looks like this:

[![Server architecture diagram. There's a Cloud Load Balancer in the front that redirects requests to Cloud Function instances in us-central1, europe-west1, asia-northeast1 regions. The functions connect to Cloud Storage in the US multi-region. Logging and Monitoring are used for monitoring. Cloud DNS is used for DNS resolutions.](images/api-diagram.png)](images/api-diagram.png)

We have the same function deployed in three regions (US, Belgium, Japan) that fetches digits from a [Cloud Storage](https://cloud.google.com/storage) bucket in the US multi-region.
The API endpoint, `api.pi.delivery` is exposed via a [Global HTTP(s) Load Balancer](https://cloud.google.com/load-balancing/docs/https). The authorative DNS servers are provided by [Cloud DNS](https://cloud.google.com/dns).

The frontend code is developed in [TypeScript](https://www.typescriptlang.org/) and [React](https://reactjs.org/).

Finally, this site is hosted on [Firebase Hosting](https://firebase.google.com/docs/hosting/), which not only automatically serves the static content via a CDN, it also automatically provisions a SSL certificate for the custom domain!

We have the entire source code published on [GitHub](https://github.com/googlecloudplatform/pi-delivery) including Terraform scripts that we use to provision the infrastructure.
