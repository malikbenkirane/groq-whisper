#!/bin/sh

# workaround for ./groq-whisper.exe dev build --push
VERSION=$(./groq-whisper.exe version)
./groq-whisper.exe dev buil
echo pushing $VERSION
./groq-whisper.exe dev build && \
				$HOME/google-cloud-sdk/gcloud cp ./groq-$VERSION gs://groq-whisper && \
				$HOME/google-cloud-sdk/gcloud cp ./setup/groq-setup-$VERSION gs://groq-whisper
