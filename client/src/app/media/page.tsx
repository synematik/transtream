"use client";

import { useEffect, useRef } from "react";

export default function VideoStream() {
    const videoRef = useRef(null);

    useEffect(() => {
        if (!window.MediaSource) {
            console.error("MediaSource API is not supported in your browser.");
            return;
        }

        const video = videoRef.current;
        const mediaSource = new MediaSource();
        video.src = URL.createObjectURL(mediaSource);

        let sourceBuffer = null;
        let ws = null;
        let segmentQueue = [];
        // Ensure the MIME type and codecs match what FFmpeg produces.
        const mimeCodec = 'video/mp4; codecs="avc1.42E01E, mp4a.40.2"';

        // Helper: Attempt to append any queued segments.
        const appendSegments = () => {
            // Do nothing if there's no sourceBuffer, it's busy, or nothing queued.
            if (!sourceBuffer || sourceBuffer.updating || segmentQueue.length === 0) {
                return;
            }
            // Check if the video element has already encountered an error.
            if (video.error) {
                console.error("Video element error detected; skipping segment append.");
                return;
            }
            const segment = segmentQueue.shift();
            try {
                sourceBuffer.appendBuffer(segment);
            } catch (err) {
                console.error("Error appending buffer:", err);
                // Optionally, you could re-queue the segment here.
            }
        };

        mediaSource.addEventListener("sourceopen", () => {
            try {
                sourceBuffer = mediaSource.addSourceBuffer(mimeCodec);
            } catch (err) {
                console.error("Error while adding SourceBuffer:", err);
                return;
            }

            sourceBuffer.mode = "segments";

            // Append any pending segments after each update.
            sourceBuffer.addEventListener("updateend", appendSegments);

            // Log any SourceBuffer errors.
            sourceBuffer.addEventListener("error", (e) => {
                console.error("SourceBuffer error:", e);
            });

            // Now that the SourceBuffer is ready, open the WebSocket connection.
            ws = new WebSocket("ws://192.168.137.137:8079/stream");
            ws.binaryType = "arraybuffer";

            ws.addEventListener("message", (event) => {
                // Check video.error before trying to append.
                if (video.error) {
                    console.error("Video element in error state; dropping segment.");
                    return;
                }
                const segment = event.data;
                if (sourceBuffer.updating || segmentQueue.length > 0) {
                    segmentQueue.push(segment);
                } else {
                    try {
                        sourceBuffer.appendBuffer(segment);
                    } catch (err) {
                        console.error("Error appending buffer:", err);
                        // In case of error, queue the segment for a later attempt.
                        segmentQueue.push(segment);
                    }
                }
            });

            ws.addEventListener("error", (e) => {
                console.error("WebSocket error:", e);
            });

            ws.addEventListener("close", () => {
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
            if (ws) {
                ws.close();
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
        <div>
            <h1>MediaSource Video Stream with Audio</h1>
            <video
                ref={videoRef}
                controls
                autoPlay
                style={{ width: "640px", height: "480px" }}
            />
        </div>
    );
}
