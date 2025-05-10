# YouTube Video Circle Bot

[Russian description](./README.RU.md)

## Overview

The **YouTube Video Circle Bot** is a Telegram bot that converts YouTube videos into circular video messages (video notes) based on provided links and time markers. Users can specify a YouTube video URL, start time, and duration to create a short video clip optimized for Telegram's video note format.

The bot is written in Go, uses FFmpeg for video processing, and is containerized using Docker for easy deployment. It supports both long polling and webhook modes for receiving Telegram updates.

## Features

- Converts YouTube videos to circular video messages.
- Supports custom start time and duration (e.g., `https://youtu.be/dQw4w9WgXcQ 00:00:43 10` for a 10-second clip starting at 43 seconds).
- Handles both full YouTube URLs and short `youtu.be` links.
- Configurable via environment variables.
- Supports proxy for YouTube API requests.
- Dockerized for easy deployment.

## Prerequisites

- **Go**: Version 1.23 or higher (for development).
- **Docker**: For building and running the containerized application.
- **FFmpeg**: Required for video processing (included in the Docker image).
- **Telegram Bot Token**: Obtain from [BotFather](https://t.me/BotFather).

## Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```

2. **Set up environment variables**:
   Create a `.env` file in the project root with the following content:
   ```
   TELEGRAM_BOT_TOKEN=your-telegram-bot-token
   WH_URL=https://your-webhook-url (optional, for webhook mode)
   PROXY_URL=http://your-http-proxy-url (optional)
   STORAGE_PATH=storage
   FFMPEG_BIN=ffmpeg
   ```
   Alternatively, set these variables in your environment.

   To run in long pooling mode, do not specify `WH_URL`. 

3. **Install dependencies**:
   ```bash
   go mod download
   ```

4. **Run locally** (without Docker):
   ```bash
   go run main.go
   ```

## Building and Running with Docker

1. **Build the Docker image**:
   ```bash
   ./build_image.sh [-y] <version>
   ```
   - `-y`: Automatically confirm all prompts.
   - `<version>`: Specify the image version (e.g., `1.0`).

   Example:
   ```bash
   ./build_image.sh -y 1.0
   ```

2. **Run the Docker container**:
   ```bash
   docker run -d --env-file .env kiruha01/yt_circles:<version>
   ```

## Usage

1. Start the bot by sending `/start` or `/help` in Telegram.
2. Send a YouTube link with optional start time and duration:
   ```
   https://youtu.be/dQw4w9WgXcQ 00:00:43 10
   ```
   - Formats supported:
     - `https://youtube.com/...` or `https://youtu.be/...`
     - Start time: `HH:MM:SS` or seconds (e.g., `43`).
     - Duration: Seconds or `HH:MM:SS` (default/max: 60 seconds).
   - Examples:
     - `https://youtu.be/dQw4w9WgXcQ 00:00:43 10`: Clip from 43s to 53s.
     - `https://youtu.be/dQw4w9WgXcQ 43`: Clip from 43s to 103s.
     - `https://youtu.be/dQw4w9WgXcQ`: Clip from 0s to 60s.

3. The bot will process the video and send a circular video note, followed by the video title.

## Project Structure

- `main.go`: Entry point for the application.
- `config/`: Configuration loading and schema.
- `telegram/`: Telegram bot logic, including long polling and webhook modes.
- `ytVideoMaker/`: YouTube video downloading and processing.
- `build_image.sh`: Script for building and pushing Docker images.
- `Dockerfile`: Multi-stage Dockerfile for building the application.
