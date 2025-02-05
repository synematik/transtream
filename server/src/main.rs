use axum::{
    extract::State,
    response::IntoResponse,
    routing::{get, post},
    Router,
};
use crossbeam_channel::{bounded, Receiver, Sender};
use gstreamer::prelude::*;
use gstreamer_app::AppSink;
use std::{
    sync::Arc,
    time::Duration,
};
use tokio::sync::broadcast;

/// Server state holding channels for GStreamer communication
#[derive(Clone)]
struct AppState {
    cmd_tx: Sender<Command>,
    frame_tx: broadcast::Sender<Vec<u8>>,
}

/// Commands to control GStreamer pipeline
#[derive(Debug)]
enum Command {
    Play,
    Pause,
    Seek(f64),
}

#[tokio::main]
async fn main() {
    // Initialize logging
    tracing_subscriber::fmt::init();

    // Create channels for pipeline control
    let (cmd_tx, cmd_rx) = bounded(10);
    let (frame_tx, _) = broadcast::channel(10);

    // Initialize GStreamer
    gstreamer::init().unwrap();

    // Start GStreamer thread
    let state = Arc::new(AppState {
        cmd_tx: cmd_tx.clone(),
        frame_tx: frame_tx.clone(),
    });
    std::thread::spawn(move || gstreamer_thread(cmd_rx, frame_tx));

    // Build Axum router
    let app = Router::new()
        .route("/ws", get(ws_handler))
        .route("/play", post(play_handler))
        .route("/pause", post(pause_handler))
        .route("/seek", post(seek_handler))
        .with_state(state);

    // Start server
    let listener = tokio::net::TcpListener::bind("0.0.0.0:3001").await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

/// WebSocket handler for streaming frames
async fn ws_handler(ws: axum::extract::ws::WebSocketUpgrade, State(state): State<Arc<AppState>>) -> impl IntoResponse {
    ws.on_upgrade(|mut socket| async move {
        let mut rx = state.frame_tx.subscribe();
        while let Ok(frame) = rx.recv().await {
            if socket.send(axum::extract::ws::Message::Binary(frame)).await.is_err() {
                break;
            }
        }
    })
}

/// Play command handler
async fn play_handler(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    state.cmd_tx.send(Command::Play).unwrap();
    "Playing"
}

/// Pause command handler
async fn pause_handler(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    state.cmd_tx.send(Command::Pause).unwrap();
    "Paused"
}

/// Seek command handler
async fn seek_handler(
    axum::extract::Query(params): axum::extract::Query<std::collections::HashMap<String, String>>,
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    if let Some(time) = params.get("time").and_then(|t| t.parse().ok()) {
        state.cmd_tx.send(Command::Seek(time)).unwrap();
        format!("Seeking to {}s", time)
    } else {
        "Invalid time parameter".to_string()
    }
}

/// GStreamer processing thread
fn gstreamer_thread(cmd_rx: Receiver<Command>, frame_tx: broadcast::Sender<Vec<u8>>) {
    let pipeline = gstreamer::parse_launch(
        "filesrc location=./test.mp4 ! decodebin ! videoconvert ! videorate ! video/x-raw,framerate=30/1 ! jpegenc quality=85 ! appsink name=appsink",
    )
        .unwrap();

    let appsink = pipeline
        .downcast_ref::<gstreamer::Bin>()
        .unwrap()
        .by_name("appsink")
        .unwrap()
        .downcast::<AppSink>()
        .unwrap();

    pipeline.set_state(gstreamer::State::Ready).unwrap();

    loop {
        // Process commands
        while let Ok(cmd) = cmd_rx.try_recv() {
            match cmd {
                Command::Play => pipeline.set_state(gstreamer::State::Playing).unwrap(),
                Command::Pause => pipeline.set_state(gstreamer::State::Paused).unwrap(),
                Command::Seek(time) => {
                    let seek_event = gstreamer::event::Seek::new(
                        1.0,
                        gstreamer::Format::Time,
                        gstreamer::SeekFlags::FLUSH | gstreamer::SeekFlags::ACCURATE,
                        gstreamer::SeekType::Set,
                        time.seconds(),
                        gstreamer::SeekType::None,
                        0,
                    );
                    pipeline.send_event(seek_event).unwrap();
                }
            }
        }

        // Process frames when playing
        if pipeline.current_state() == gstreamer::State::Playing {
            if let Ok(sample) = appsink.pull_sample(Some(gstreamer::ClockTime::from_millis(33))) {
                if let Some(buffer) = sample.buffer() {
                    let map = buffer.map_readable().unwrap();
                    let _ = frame_tx.send(map.as_slice().to_vec());
                }
            }
        } else {
            std::thread::sleep(Duration::from_millis(100));
        }
    }
}