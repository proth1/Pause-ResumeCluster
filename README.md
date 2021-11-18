# Pausing and Resuming CastAI Managed Kubernetes Cluster

## AWS Serverless Setup for Pause/Resume
The AWS Pause/Resume schedule uses a simple Lambda Function, EventBridge jobs and Secret Manager to easily automate cluster pause/resume. 

[Setup Pause/Resume in AWS](./pause-resume-aws/README.md)

## GCP Serverless Setup for Pause/Resume
The GCP Pause/Resume schedule uses a Cloud Function, Cloud Scheduler and Secret Manager to automate the pause/resume of your GKE clusters managed by CastAI. 

[Setup Pause/Resume in GCP](./pause-resume-gcp/README.md)