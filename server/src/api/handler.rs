use crate::config::Config;
use crate::ffmpeg::FFmpeg;
use crate::model::{MovieTable, Table};
use crate::sqlite::SharedDb;
use hyper::{header, Body, Response, StatusCode};
use std::sync::Arc;

pub async fn get_stream(
    config: Arc<Config>,
) -> Result<Response<Body>, hyper::Error> {
    let config = Arc::new(config.ffmpeg.clone());
    let ffmpeg = FFmpeg::new(config);
    let (tx, body) = Body::channel();

    tokio::spawn(async move {
        ffmpeg.transcode("Marked for Death (1990).mkv", tx).await;
    });

    let resp = Response::builder()
        .header("Content-Type", "video/mp4")
        .header("Content-Disposition", "inline")
        .header("Content-Transfer-Enconding", "binary")
        .body(body)
        .unwrap();

    Ok(resp)
}
