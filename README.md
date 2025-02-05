[manifest.m3u8](https://synema.cxllmerichie.com/proxy/0c56f96f6f49434d110f8a0a98737443:2025020601:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFJuRVplSE9qeTA1Zmx3anhPRG4rOHc9PQ==/2/4/8/6/0/4/f0yob.mp4:hls:manifest.m3u8)

# RTMP
1. [RTMP](https://hub.docker.com/r/tiangolo/nginx-rtmp/)
2. ffmpeg -re -i "https://synema.cxllmerichie.com/proxy/0c56f96f6f49434d110f8a0a98737443:2025020601:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFJuRVplSE9qeTA1Zmx3anhPRG4rOHc9PQ==/2/4/8/6/0/4/f0yob.mp4:hls:manifest.m3u8" -vf scale=1280:960 -vcodec libx264 -profile:v baseline -pix_fmt yuv420p -f flv rtmp://127.0.0.1/live/streamid
3. [RTMP2HLS](https://stackoverflow.com/questions/19658216/how-can-we-transcode-live-rtmp-stream-to-live-hls-stream-using-ffmpeg)

# WebSocket
ffmpeg -re -i "https://synema.cxllmerichie.com/proxy/0c56f96f6f49434d110f8a0a98737443:2025020601:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFJuRVplSE9qeTA1Zmx3anhPRG4rOHc9PQ==/2/4/8/6/0/4/f0yob.mp4:hls:manifest.m3u8" -c:v libx264 -preset ultrafast -tune zerolatency -c:a aac -f mpegts "ws://localhost:8080"

# RTSP [docker](https://github.com/bluenviron/mediamtx?tab=readme-ov-file#ffmpeg)
1. docker run --rm -it --network=host bluenviron/mediamtx:latest
    ```
    docker run --rm -it \
    -e MTX_RTSPTRANSPORTS=tcp \
    -e MTX_WEBRTCADDITIONALHOSTS=192.168.x.x \
    -p 8554:8554 \
    -p 1935:1935 \
    -p 8888:8888 \
    -p 8889:8889 \
    -p 8890:8890/udp \
    -p 8189:8189/udp \
    bluenviron/mediamtx 
    ```
2. ffmpeg -re -stream_loop -1 -i https://synema.cxllmerichie.com/proxy/0c56f96f6f49434d110f8a0a98737443:2025020601:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFJuRVplSE9qeTA1Zmx3anhPRG4rOHc9PQ==/2/4/8/6/0/4/f0yob.mp4:hls:manifest.m3u8 -c copy -f rtsp -rtsp_transport tcp rtsp://localhost:8554/stream

ffmpeg -re -listen -1 -i https://synema.cxllmerichie.com/proxy/0c56f96f6f49434d110f8a0a98737443:2025020601:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFJuRVplSE9qeTA1Zmx3anhPRG4rOHc9PQ==/2/4/8/6/0/4/f0yob.mp4:hls:manifest.m3u8 -c copy -f rtsp -rtsp_transport tcp rtsp://localhost:8554/stream

# FFPLAY
1. ffplay -rtsp_transport tcp rtsp://localhost:8554/stream