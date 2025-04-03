"use client";

import React from "react";

export default function VideoStream() {
    const videoRef = React.useRef<HTMLVideoElement>(null);

    React.useEffect(() => {
        if (!window.MediaSource) {
            console.error("MediaSource API is not supported in your browser.");
            return;
        }

        const video = videoRef.current!;
        const mediaSource = new MediaSource();
        video.src = URL.createObjectURL(mediaSource);

        let buffer: SourceBuffer;
        let socket: WebSocket;
        let segments: Array<BufferSource> = [];
        // Set MIME type for WebM output.
        const mimeCodec = 'video/webm; codecs="vp8, vorbis"';

        const appendSegments = () => {
            if (!buffer || buffer.updating || segments.length === 0) {
                return;
            }
            if (video.error) {
                console.error("Video element in error state; dropping segment.");
                return;
            }
            const segment = segments.shift();
            try {
                buffer.appendBuffer(segment);
            } catch (err) {
                console.error("Error appending buffer:", err);
                segments.push(segment);
            }
        };

        mediaSource.addEventListener("sourceopen", () => {
            try {
                buffer = mediaSource.addSourceBuffer(mimeCodec);
            } catch (err) {
                console.error("Error while adding SourceBuffer:", err);
                return;
            }

            buffer.mode = "segments";
            buffer.addEventListener("updateend", appendSegments);
            buffer.addEventListener("error", (e) => {
                console.error("SourceBuffer error:", e);
            });

            socket = new WebSocket("ws://127.0.0.1:8079/socket");
            socket.binaryType = "arraybuffer";

            socket.addEventListener("message", (event) => {
                if (video.error) {
                    console.error("Video element in error state; dropping segment.");
                    return;
                }
                const segment: BufferSource = event.data;
                if (buffer.updating || segments.length > 0) {
                    segments.push(segment);
                } else {
                    try {
                        buffer.appendBuffer(segment);
                    } catch (err) {
                        console.error("Error appending buffer:", err);
                        segments.push(segment);
                    }
                }
            });

            socket.addEventListener("error", (e) => {
                console.error("WebSocket error:", e);
            });

            socket.addEventListener("close", () => {
                if (mediaSource.readyState === "open") {
                    try {
                        mediaSource.endOfStream();
                    } catch (err) {
                        console.error("Error ending media source:", err);
                    }
                }
            });
        });

        video.addEventListener("error", (e) => {
            console.error("Video element error:", video.error);
        });

        return () => {
            if (socket) {
                socket.close();
            }
            if (mediaSource.readyState === "open") {
                try {
                    mediaSource.endOfStream();
                } catch (err) {
                    console.error("Error during mediaSource.endOfStream:", err);
                }
            }
        };
    }, []);

    return (
        <video ref={videoRef} controls autoPlay muted className={'size-full'} />
    );
}
