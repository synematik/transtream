use axum::{Router, Extension, response::IntoResponse};
use axum::http::StatusCode;
use ffmpeg_next::{codec, format, media, packet, software, util};
use std::sync::{Arc, RwLock};
use tokio::sync::RwLock as TokioRwLock;
use tokio::io::{AsyncReadExt, AsyncWriteExt};

#[derive(Debug, Clone, Copy, PartialEq)]
enum StreamState {
    Play,
    Pause,
}

struct Gateway {
    state: Arc<TokioRwLock<StreamState>>,
    video_decoder: Option<ffmpeg_next::decoder::Video>,
}

impl Gateway {
    fn new() -> Self {
        Gateway {
            state: Arc::new(TokioRwLock::new(StreamState::Play)),
            video_decoder: None,
        }
    }

    async fn set_state(&self, state: StreamState) {
        let mut lock = self.state.write().await;
        *lock = state;
    }

    async fn get_state(&self) -> StreamState {
        let lock = self.state.read().await;
        *lock
    }

    // Initializes FFmpeg for decoding
    async fn init_ffmpeg(&mut self) -> Result<(), String> {
        ffmpeg_next::init().map_err(|e| format!("Failed to initialize FFmpeg: {}", e))?;

        // Open input file (could also be a stream)
        let mut context = format::input(&"input_video.mp4").map_err(|e| format!("Failed to open input: {}", e))?;

        // Find the video stream (assuming one video stream in the file)
        let input_stream = context.streams().best(media::Type::Video).ok_or("No video stream found")?;
        let codec = input_stream.codec().decoder().video().ok_or("Failed to find video decoder")?;

        self.video_decoder = Some(codec);

        Ok(())
    }

    // Handles streaming video frames
    async fn stream_video(&self, response: axum::response::Response) {
        // Get the state of the stream
        let state = self.get_state().await;

        if state == StreamState::Play {
            // Decode and stream the frames
            if let Some(decoder) = &self.video_decoder {
                let mut packet = packet::Packet::empty();
                while decoder.decode(&mut packet).is_ok() {
                    if packet.is_empty() {
                        continue;
                    }

                    // You can process the frame here if needed, e.g., scale or modify it
                    // Send the decoded frame to the client
                    let frame = decoder.decode_packet(&packet);
                    if let Ok(frame) = frame {
                        // Send the frame to the HTTP stream
                        let buffer = frame.data();
                        response.write_all(&buffer).await.unwrap();
                    }
                }
            }
        } else {
            println!("Stream is paused. Holding frames.");
        }
    }
}

// API endpoint to control pause/play state
async fn play(Extension(gateway): Extension<Gateway>) -> StatusCode {
    gateway.set_state(StreamState::Play).await;
    StatusCode::OK
}

async fn pause(Extension(gateway): Extension<Gateway>) -> StatusCode {
    gateway.set_state(StreamState::Pause).await;
    StatusCode::OK
}

#[tokio::main]
async fn main() {
    let gateway = Gateway::new();

    // Initialize FFmpeg decoder
    match gateway.init_ffmpeg().await {
        Ok(()) => println!("FFmpeg initialized successfully"),
        Err(e) => panic!("Error initializing FFmpeg: {}", e),
    }

    // Set up Axum server
    let app = Router::new()
        .route("/api/control/play", axum::routing::get(play))
        .route("/api/control/pause", axum::routing::get(pause))
        .route("/stream", axum::routing::get(move || {
            gateway.stream_video(); // Start streaming frames
            StatusCode::OK
        }))
        .layer(Extension(gateway));

    axum::Server::bind(&"0.0.0.0:8000".parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}
