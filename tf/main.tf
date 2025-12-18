terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.14.1"
    }
  }

  backend "gcs" {
    bucket = "groq-whisper"
  }
}

provider "google" {
  project = var.project
  region  = var.region
}
