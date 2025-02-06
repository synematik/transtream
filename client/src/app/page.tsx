"use client";

import React from 'react';
import Hls from 'hls.js';

export default function PlayerPage({}) {
    const streamId = "fixme:hardcoded";

    const videoRef = React.useRef<HTMLVideoElement>(null);
    const [isPlaying, setIsPlaying] = React.useState(false);
    const [position, setPosition] = React.useState(0);
    const ws = React.useRef<WebSocket>(null);

    React.useEffect(() => {
        const hls = new Hls();
        hls.loadSource(`http://192.168.137.137:8080/stream/${streamId}/manifest.m3u8`);
        hls.attachMedia(videoRef.current);

        // WebSocket connection
        ws.current = new WebSocket(`ws://192.168.137.137:8080/ws/${streamId}`);
        ws.current.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'state') {
                setIsPlaying(data.playing);
                // Sync if discrepancy > 0.5s
                if (Math.abs(videoRef.current.currentTime - data.position) > 0.5) {
                    videoRef.current.currentTime = data.position;
                }
            }
        };

        return () => {
            hls.destroy();
            ws.current.close();
        };
    }, [streamId]);

    // Control handlers
    const handlePlay = async () => fetch(`http://192.168.137.137:8080/api/play/${streamId}`, {method: 'POST'});
    const handlePause = async () => fetch(`http://192.168.137.137:8080/api/pause/${streamId}`, {method: 'POST'});
    const handleSeek = async (time) => fetch(`http://192.168.137.137:8080/api/seek/${streamId}?time=${time}`, {method: 'POST'});

    return (
        <div className="player-container">
            <video
                ref={videoRef}
                onPlay={handlePlay}
                onPause={handlePause}
                onSeeked={(e) => handleSeek(e.currentTarget.currentTime)}
                controls
            />
            <div className="controls">
                <button onClick={isPlaying ? handlePause : handlePlay}>
                    {isPlaying ? '⏸️ Pause' : '▶️ Play'}
                </button>
                <input
                    type="range"
                    min="0"
                    max={videoRef.current?.duration || 100}
                    value={position}
                    onChange={handleSeek}
                />
            </div>
        </div>
    );
}