"use client";

import React from 'react';

const streamURL = 'https://synema.cxllmerichie.com/proxy/11c92a546075985695d11d4a7dbec3e3:2025020619:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydHNRa3Y2R053b3h5cGFDd21VYlUvaGc9PQ==/2/4/8/6/2/3/i8dqz.mp4:hls:manifest.m3u8';

const Page: React.FC<{}> = ({}) => {
    const canvasRef = React.useRef(null);
    const [ws, setWs] = React.useState(null);

    React.useEffect(() => {
        // Connect to Media Server
        const ws = new WebSocket(`ws://127.0.0.1:8080/${btoa(streamURL)}`);
        setWs(ws);

        // Initialize canvas context
        const canvas = canvasRef.current;
        const ctx = canvas.getContext('2d');

        // Handle incoming JPEG frames
        ws.onmessage = (event) => {
            const blob = new Blob([event.data], {type: 'image/jpeg'});
            const url = URL.createObjectURL(blob);
            const img = new Image();

            console.log(event)
            img.onload = () => {
                ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
                URL.revokeObjectURL(url);
            };

            img.src = url;
        };

        // Send player events
        const sendEvent = (type, value) => {
            ws.send(JSON.stringify({
                type,
                ...(value !== undefined && {value})
            }));
        };

        // Attach event listeners for play/pause/seek
        canvas.addEventListener('click', () => {
            sendEvent('play');
        });

        return () => {
            ws.close();
        };
    }, []);

    return (
        <canvas
            ref={canvasRef}
            width="1280"
            height="720"
            className={'border'}
        />
    );
}

export default Page;