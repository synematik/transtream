"use client";

import React from "react";

export default function VideoStream() {
    const canvasRef = React.useRef<HTMLCanvasElement>(null);
    const width = 640;
    const height = 480;

    React.useEffect(() => {
        // Connect to the WebSocket endpoint on your Go server.
        const ws = new WebSocket("ws://192.168.137.137:8079/stream");
        // Expect binary data as an ArrayBuffer.
        ws.binaryType = "arraybuffer";

        ws.onmessage = (event) => {
            const arrayBuffer = event.data;
            const canvas = canvasRef.current;
            if (!canvas) return;
            const ctx = canvas.getContext("2d")!;

            // The received ArrayBuffer should be exactly width*height*4 bytes.
            const imageDataArray = new Uint8ClampedArray(arrayBuffer);
            // Create an ImageData object from the raw RGBA bytes.
            const imageData = new ImageData(imageDataArray, width, height);
            // Draw the ImageData on the canvas.
            ctx.putImageData(imageData, 0, 0);
        };

        ws.onerror = (error) => {
            console.error("WebSocket error:", error);
        };

        // Clean up the WebSocket on component unmount.
        return () => {
            ws.close();
        };
    }, []);

    return (
        <canvas
            ref={canvasRef}
            width={width}
            height={height}
        />
    );
}