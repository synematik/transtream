import React from 'react';

export default function Page() {
    const canvasRef = React.useRef<HTMLCanvasElement>(null);
    const [ws, setWs] = React.useState<WebSocket | null>(null);

    React.useEffect(() => {
        const newWs = new WebSocket('ws://localhost:3001/ws');
        newWs.binaryType = 'arraybuffer';

        newWs.onmessage = (event) => {
            const blob = new Blob([event.data], {type: 'image/jpeg'});
            const url = URL.createObjectURL(blob);
            const img = new Image();

            img.onload = () => {
                const ctx = canvasRef.current?.getContext('2d');
                if (ctx) {
                    ctx.drawImage(img, 0, 0, ctx.canvas.width, ctx.canvas.height);
                }
                URL.revokeObjectURL(url);
            };

            img.src = url;
        };

        setWs(newWs);
        return () => newWs.close();
    }, []);

    const sendCommand = async (command: string) => {
        await fetch(`http://localhost:3001/${command}`, {method: 'POST'});
    };

    const seek = async (time: number) => {
        await fetch(`http://localhost:3001/seek?time=${time}`, {method: 'POST'});
    };

    return (
        <div>
            <canvas
                ref={canvasRef}
                width="1280"
                height="720"
                style={{border: '1px solid black'}}
            />
            <div>
                <button onClick={() => sendCommand('play')}>Play</button>
                <button onClick={() => sendCommand('pause')}>Pause</button>
                <button onClick={() => seek(30)}>Seek to 30s</button>
            </div>
        </div>
    );
}