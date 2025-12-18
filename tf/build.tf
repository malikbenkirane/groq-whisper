resource "google_cloudbuild_trigger" "build" {
  location    = var.region
  project     = var.project
  description = "Build Trigger for groq-whisper on tags"

  github {
    owner = "malikbenkirane"
    name  = "groq-whisper"

    push {
      tag = "v.*"
    }
  }

  build {
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "build",
        "-t", local.image,
        "-f", "./deploy/Dockerfile",
        "./deploy"
      ]
    }
    images = [
      local.image
    ]
  }
}

locals {
  image = "${var.image_host}/${var.project}/groq-whisper/deploy:$TAG_NAME"
}
