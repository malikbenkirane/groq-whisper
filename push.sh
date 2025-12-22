#!/bin/sh

# workaround for ./groq-whisper.exe dev build --push
VERSION=$(./groq-whisper.exe version)
echo pushing $VERSION
./groq-whisper.exe dev build && \
				gcloud cp ./groq-$VERSION gs://groq-whisper && \
				gcloud cp ./setup/groq-setup-$VERSION gs://groq-whisper
