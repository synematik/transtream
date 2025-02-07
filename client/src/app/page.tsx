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
        // Ensure the MIME type and codecs match what FFmpeg produces.
        const codec = 'video/mp4; codecs="avc1.42E01E, mp4a.40.2"';

        // Helper: Attempt to append any queued segments.
        const appendSegments = () => {
            // Do nothing if there's no buffer, it's busy, or nothing queued.
            if (!buffer || buffer.updating || segments.length === 0) {
                return;
            }
            // Check if the video element has already encountered an error.
            if (video.error) {
                console.error("Video element error detected; skipping segment append.");
                return;
            }
            const segment = segments.shift();
            try {
                buffer.appendBuffer(segment);
            } catch (err) {
                console.error("Error appending buffer:", err);
                // Optionally, you could re-queue the segment here.
            }
        };

        mediaSource.addEventListener("sourceopen", () => {
            try {
                buffer = mediaSource.addSourceBuffer(codec);
            } catch (err) {
                console.error("Error while adding SourceBuffer:", err);
                return;
            }

            buffer.mode = "segments";

            // Append any pending segments after each update.
            buffer.addEventListener("updateend", appendSegments);

            // Log any SourceBuffer errors.
            buffer.addEventListener("error", (e) => {
                console.error("SourceBuffer error:", e);
            });

            // Now that the SourceBuffer is ready, open the WebSocket connection.
            socket = new WebSocket("ws://192.168.137.137:8079");
            socket.binaryType = "arraybuffer";

            socket.addEventListener("message", (event) => {
                // Check video.error before trying to append.
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
                        // In case of error, queue the segment for a later attempt.
                        segments.push(segment);
                    }
                }
            });

            socket.addEventListener("error", (e) => {
                console.error("WebSocket error:", e);
            });

            socket.addEventListener("close", () => {
                // Signal end-of-stream if the MediaSource is still open.
                if (mediaSource.readyState === "open") {
                    try {
                        mediaSource.endOfStream();
                    } catch (err) {
                        console.error("Error ending media source:", err);
                    }
                }
            });
        });

        // Listen for errors on the video element itself.
        video.addEventListener("error", (e) => {
            console.error("Video element error:", video.error);
        });

        // Cleanup on unmount.
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
        <video
            ref={videoRef}
            controls
            autoPlay
            muted
            className={'size-full'}
        />
    );
}
