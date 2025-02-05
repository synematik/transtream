use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use ffmpeg_next::{codec, format, media, software, packet, error};
use std::fs::File;
use std::io::{self, Read};

async fn stream_video() -> impl Responder {
    // Open the video file using ffmpeg-next
    let file_path = "video.mp4";
    let mut input = format::input(file_path).unwrap();

    // Initialize the decoder (video stream)
    let mut stream = input.streams().best(media::Type::Video).unwrap();
    let decoder = codec::context::Context::from_stream(&stream).unwrap();
    let mut decoder = software::decoder::Video::new(decoder).unwrap();

    // Prepare for reading the video frames
    let mut packet = packet::Packet::empty();
    let mut frame_buffer = Vec::new();

    // Create an HTTP response stream
    let response = HttpResponse::Ok()
        .content_type("image/jpeg") // Sending frames as JPEG images
        .streaming();

    let mut res_stream = response.await?;

    // Read frames and send them one by one
    while let Ok(true) = input.read_packet(&mut packet) {
        if packet.stream() == stream.index() {
            // Decode the video frame
            if let Ok(frame) = decoder.decode(&packet) {
                frame_buffer.clear();

                // Convert the frame to JPEG (or any other format you want)
                let mut jpg_encoder = image::jpeg::JPEGEncoder::new(&mut frame_buffer);
                jpg_encoder.encode(&frame.data(0), frame.width() as u32, frame.height() as u32, image::ColorType::Rgb8).unwrap();

                // Send the JPEG frame
                res_stream.send_data(frame_buffer.clone()).await.unwrap();
            }
        }
    }

    res_stream.await?;
    HttpResponse::Ok().finish() // Close connection after streaming is done
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .route("/video", web::get().to(stream_video)) // Stream video on /video endpoint
    })
        .bind("127.0.0.1:8080")?
        .run()
        .await
}
