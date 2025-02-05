import asyncio
import subprocess
from collections import defaultdict
import base64

import uvicorn
from fastapi import FastAPI, WebSocket

app = FastAPI()
active_streams = defaultdict(lambda: {
    'clients': set(),
    'process': None,
    'is_playing': True,
    'current_time': 0.0
})


async def stream(stream_id: str, manifest_url: str):
    cmd = [
        'ffmpeg',
        '-i', manifest_url,
        '-f', 'image2pipe',
        '-c:v', 'mjpeg',
        '-q:v', '5',
        '-'
    ]

    proc = await asyncio.create_subprocess_exec(
        *cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE
    )

    active_streams[stream_id]['process'] = proc
    active_streams[stream_id]['is_playing'] = True

    try:
        while True:
            if active_streams[stream_id]['is_playing']:
                # Read a JPEG frame from FFmpeg
                frame_size_bytes = await proc.stdout.read(4)
                if not frame_size_bytes:
                    break

                frame_size = int.from_bytes(frame_size_bytes, byteorder='big')
                frame_data = await proc.stdout.read(frame_size)

                # Broadcast frame to all clients
                for ws in active_streams[stream_id]['clients']:
                    try:
                        await ws.send_bytes(frame_data)
                    except:
                        continue

                # Update current time (simplified)
                active_streams[stream_id]['current_time'] += 1 / 30  # Assuming 30 FPS
            else:
                # Pause: Stop reading frames
                await asyncio.sleep(0.1)

    finally:
        proc.kill()
        del active_streams[stream_id]


@app.websocket("/{manifest_url}")
async def _(websocket: WebSocket, manifest_url: str):
    manifest_url = base64.b64decode(manifest_url).decode("utf-8")
    await websocket.accept()

    stream_id = 'fixme:hardcoded'
    active_streams[stream_id]['clients'].add(websocket)

    # Start stream if first client
    if not active_streams[stream_id]['process']:
        asyncio.create_task(stream(stream_id, manifest_url))

    try:
        while True:
            data = await websocket.receive_json()
            if data['type'] == 'pause':
                active_streams[stream_id]['is_playing'] = False
            elif data['type'] == 'play':
                active_streams[stream_id]['is_playing'] = True
            elif data['type'] == 'seek':
                # Kill current process and restart from seek position
                active_streams[stream_id]['process'].kill()
                active_streams[stream_id]['current_time'] = data['time']
                asyncio.create_task(stream(stream_id, manifest_url))

    except Exception as e:
        print(e)
        active_streams[stream_id]['clients'].remove(websocket)


if __name__ == '__main__':
    uvicorn.run(app, host='127.0.0.1', port=8080)
